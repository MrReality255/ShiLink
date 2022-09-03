/*
	ShiLink Server creates a server translating its requests into cobra CLI commands.
	Requests are required to be POSTs, commands are parsed from the URL, flags from the request body.

	Example:
	http://localhost:port/api			{"help":true}
	http://localhost:port/api/contract/list
*/

package shilink

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type shiLinkServer struct {
	cmd *cobra.Command
}

func (ss *shiLinkServer) buildCommandLine(req *http.Request) ([]string, error) {
	// split the URL into separate commands
	commands := strings.Split(req.URL.Path, "/")
	if len(commands) > 0 {
		// remove leading empty command and evtl. "api"
		if commands[0] == "" {
			commands = commands[1:]
		}

		if commands[0] == "api" {
			commands = commands[1:]
		}
	}

	// read the body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// parse flags
	if len(body) != 0 {
		flags := make(map[string]interface{})
		if err = json.Unmarshal(body, &flags); err != nil {
			return nil, err
		}

		for name, value := range flags {
			// single-letter flags start with -, longer flags start with --
			prefix := "-"
			if len(name) != 1 {
				prefix = "--"
			}
			commands = append(commands, fmt.Sprintf("%v%v", prefix, name))
			if _, isBool := value.(bool); !isBool {
				commands = append(commands, fmt.Sprintf("%v", value))
			}
		}
	}
	return commands, nil
}

func (ss *shiLinkServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// we allow only POST
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// now we parse the command line
	args, err := ss.buildCommandLine(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("received command line: %v\n", args)
	ss.cmd.ResetFlags()
	ss.cmd.SetArgs(args)

	// create an OS pipe and redirect the standard output
	origStdOut := os.Stdout

	defer func() { os.Stdout = origStdOut }()
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		fmt.Printf("unable to create a pipe: %v\n", err)
		return
	}
	os.Stdout = pipeWriter

	// execute the original command with manufactured command line arguments
	if err := ss.cmd.Execute(); err != nil {
		fmt.Printf("err: %v", err)
	}

	// collect the standard output content
	if err = pipeWriter.Close(); err != nil {
		os.Stdout = origStdOut
		fmt.Printf("unable to write the pipe writer: %v\n", err)
		return
	}
	output, err := ioutil.ReadAll(pipeReader)
	os.Stdout = origStdOut
	if err != nil {
		fmt.Printf("unable to read the content of std out: %v\n", err)
		return
	}

	// send the response back
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(output); err != nil {
		fmt.Printf("unable to respond: %v\n", err)
	}
}

func UseShiLink(rootCmd *cobra.Command) {
	var (
		port int
	)

	newCmd := *rootCmd

	server := &cobra.Command{
		Use:   "server",
		Short: "starts the ShiLink Server",
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("Starting server on port %v\n", port)
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), &shiLinkServer{&newCmd}); err != nil {
				fmt.Printf("Unable to start the server: %v\n", err)
			}
		},
	}
	server.Flags().IntVarP(&port, "port", "p", 8111, "port of the ShiLink server")
	rootCmd.AddCommand(server)
}
