// Pages 289-290
// Listing 12-23: Initial gRPC client code for our housework application
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"Ch12/housework/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var addr, caCertFn string

func init() {
	// Aside from all the new imports, you add flags for the gRPC server address
	// and its certificate.
	flag.StringVar(&addr, "address", "localhost:34443",
		"server address")
	flag.StringVar(&caCertFn, "ca-cert", "cert.pem", "CA certificate")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [flags] [add chore, ...|complete #]
add         add comma-separated chores
complete    complete designated chore

Flags:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

// Page 288
// Listing 12-24: Using the gRPC client to list the current chores
func list(ctx context.Context, client housework.RobotMaidClient) error {
	// This code is similar to Listing 12-5, except your asking the gRPC client
	// for the list of chores, which retrieves them from the server.
	// You need to pass an empty message to make gRPC happy.
	chores, err := client.List(ctx, new(housework.Empty))
	if err != nil {
		return err
	}

	if len(chores.Chores) == 0 {
		fmt.Println("You have nothing to do!")
		return nil
	}

	fmt.Println("#\t[X]\tDescription")
	for i, chore := range chores.Chores {
		c := " "
		if chore.Complete {
			c = "X"
		}
		fmt.Printf("%d\t[%s]\t%s\n", i+1, c, chore.Description)
	}

	return nil
}

// Page 291
// Listing 12-25: Adding new chores using the gRPC client
func add(ctx context.Context, client housework.RobotMaidClient, s string) error {
	chores := new(housework.Chores)

	// You parse the comma-separated list of chores.
	for _, chore := range strings.Split(s, ",") {
		if desc := strings.TrimSpace(chore); desc != "" {
			chores.Chores = append(chores.Chores, &housework.Chore{
				Description: desc,
			})
		}
	}

	// Instead of flushing these chores to disk, you pass them along to the gRPC
	// client.
	// The gRPC client transparently sends them to the gRPC server and returns
	// the response to you.
	// Since you know Rosie returns a non-nil error when the Add call fails, you
	// return the error as the result of the add function.
	var err error
	if len(chores.Chores) > 0 {
		_, err = client.Add(ctx, chores)
	}

	return err
}

// Pages 291
// Listing 12-26: Marking chores complete by using the gRPC client
func complete(ctx context.Context, client housework.RobotMaidClient, s string) error {
	// Notice the protoc-gen-go module, which converts the snake-cased
	// chore_number field in Listing 12-20 to Pascal case in the generated Go
	// code.
	// You must also convert the int returned by strconv.Atoi to an int32 before
	// assigning it to the complete request message's chore number since
	// ChoreNumber is an int32.
	i, err := strconv.Atoi(s)
	if err == nil {
		_, err = client.Complete(ctx, &housework.CompleteRequest{ChoreNumber: int32(i)})
	}

	return err
}

// Page 292
// Listing 12-27: Creating a new gRPC connection using TLS and certificate pinning
func main() {
	flag.Parse()

	caCert, err := ioutil.ReadFile(caCertFn)
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal("failed to add certificate from pool")
	}

	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences: []tls.CurveID{tls.CurveP256},
					MinVersion:       tls.VersionTLS12,
					RootCAs:          certPool,
				},
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Page 293
	// Listing: 12-28: Instantiating a new gRPC client and making calls
	rosie := housework.NewRobotMaidClient(conn)
	ctx := context.Background()

	switch strings.ToLower(flag.Arg(0)) {
	case "add":
		err = add(ctx, rosie, strings.Join(flag.Args()[1:], " "))
	case "complete":
		err = complete(ctx, rosie, flag.Arg(1))
	}

	if err != nil {
		log.Fatal(err)
	}

	err = list(ctx, rosie)
	if err != nil {
		log.Fatal(err)
	}
}
