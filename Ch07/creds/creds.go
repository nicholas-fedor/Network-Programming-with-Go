// Page 157
// Listing 7-14: Expecting group names on the command line.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

	"github.com/nicholas-fedor/Network-Programming-with-Go/Ch07/creds/auth"
)

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(
			flag.CommandLine.Output(),
			// Our application expects a series of group names as arguments.
			// You'll add the group ID for each group name to the map of allowed groups.
			"Usage:\n\t%s <group names>\n",
			filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

// Page 157
// Listing 7-15: Parsing group names into Group IDs.
func parseGroupNames(args []string) map[string]struct{} {
	groups := make(map[string]struct{})

	for _, arg := range args {
		// Retrieves the group information for each group name.
		grp, err := user.LookupGroup(arg)
		if err != nil {
			log.Println(err)
			continue
		}

		// Inserts each group ID into the groups map.
		groups[grp.Gid] = struct{}{}
	}

	return groups
}

// Page 158
// Listing 7-16: Authorizing peers based on their credentials.
func main() {
	flag.Parse()

	groups := parseGroupNames(flag.Args())
	socket := filepath.Join(os.TempDir(), "creds.sock")
	addr, err := net.ResolveUnixAddr("unix", socket)
	if err != nil {
		log.Fatal(err)
	}

	s, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	// Since we execute this service on the command line, we'll stop the service
	// by sending an interrupt signal, usually with the CTRL-C key combination.
	// However, this signal abruptly terminates the service before Go has a
	// chance to clean up the socket file, despite our use of net.ListenUnix.
	// Therefore, we need to listen for this signal.
	signal.Notify(c, os.Interrupt)
	// Then spin off a goroutine in which we gracefully close the listener after
	// receiving the signal.
	// This will ensure Go properly cleans up the socket file.
	go func ()  {
		<-c
		_ = s.Close()
	}()

	fmt.Printf("Listening on %s ...\n", socket)

	for {
		// The listener accepts connections by using AcceptUnix so a
		// *net.UnixConn is returned of the usual net.Conn, since our
		// auth.Allowed function requires a *net.UnixConn type as its first argument.
		conn, err := s.AcceptUnix()
		if err != nil {
			break
		}
		// We then determine whether the peer's credentials are allowed.
		// Allowed peers stay connected.
		// Disallowed peers are immediately disconnected.
		if auth.Allowed(conn, groups) {
			_, err := conn.Write([]byte("Welcome\n"))
			if err == nil {
				// handle the connection in a goroutine here
				continue
			}
		} else {
			_, err = conn.Write([]byte("Access Denied\n"))
		}
		if err != nil {
			log.Println(err)
		}
		_ = conn.Close()
	}
}