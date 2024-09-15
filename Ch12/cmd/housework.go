// Pages 269-284
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// "Ch12/housework"
	"Ch12/housework/v1"

	// Since the ultimate purpose of this application is to demonstrate data
	// serialization, you'll use multiple serialization formats to store the
	// data.
	// This should show you how easy it is to switch between various formats.
	// To prepare for that, you include import statements for those formats.
	// This will make it easier for you to switch between the formats later.
	// storage "Ch12/json"
	// storage "Ch12/gob"
	storage "Ch12/protobuf"
)

// Pages 171-172
// Listing 12-2: Initial housework application code
var dataFile string

func init() {
	flag.StringVar(&dataFile, "file", "housework.db", "data file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			// This sets up the command line arguments and their usage.
			// You can specify the add argument, followed by a
			// command-line separated list of chores to add to the list,
			// or you can pass the argument complete and a chore number
			// to mark the chore as complete.
			// In the absence of command line options, the app will
			// display the current list of chores.
			`
		Usage: %s add chore1, chore2, etc... | complete [chore #]
		Flags:
		add		add comma-separated chores
				Example: add Mop floors, Clean Dishes, Mow the lawn
		complete	mark designated chore number as complete (only one chore per command)
				Example: complete 1

		`, filepath.Base(os.Args[0]))
	}
}

// Page 272
// Listing 12-3: Deserializing chores from a file
// This function returns a slice of pointers to housework Chore structs from
// Listing 12-1.
func load() ([]*housework.Chore, error) {
	// If the data file does not exist, you exit early, returning an empty
	// slice.
	// This default case will occur when you first run the application.
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return make([]*housework.Chore, 0), nil
	}

	// If the application finds a data file, you open it...
	df, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := df.Close(); err != nil {
			fmt.Printf("closing data file: %v", err)
		}
	}()

	// ... and pass it along to the storage's Load function, which expects an io.Reader.
	return storage.Load(df)
}

// Page 273
// Listing 12-4: Flushing chores to the storage.
// Flush defines a function that flushes the chores in memory to your storage
// for persistence.
func flush(chores []*housework.Chore) error {
	// Here, you create a new file or truncate the existing file ...
	df, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := df.Close(); err != nil {
			fmt.Printf("closing data file: %v", err)
		}
	}()

	// ... and pass the file pointer and slice of chores to the storage's
	// Flush function.
	// This function accepts an io.Writer and your slice.
	// There's certainly room for improvement in the way you handle the existing
	// serialized file.
	// But for demonstration purposes, this will suffice.
	return storage.Flush(df, chores)
}

// Pages 273-274
// Listing 12-5: Printing the list of chores to standard output.
// This function adds functionality to display the chores on the command line.
func list() error {
	// First, you load the list of chores from storage.
	// If there are no chores in your list, you simply print as much to standard
	// output.
	// Otherwise, you print a header and the list of chores.
	chores, err := load()
	if err != nil {
		return err
	}

	if len(chores) == 0 {
		fmt.Println("You're all caught up!")
		return nil
	}

	fmt.Println("#\t[X]\tDescription")
	for i, chore := range chores {
		c := " "
		if chore.Complete {
			c = "X"
		}
		fmt.Printf("%d\t[%s]\t%s\n", i+1, c, chore.Description)
	}

	return nil
}

// Page 274
// Listing 12-7: Adding chores to the list.
func add(s string) error {
	// You retrieve the chores from storage, modify them, ...
	chores, err := load()
	if err != nil {
		return err
	}

	// You want the option to add more than one chore at a time, so you split
	// the incoming chore description by commas and append each chore to the
	// slice.
	// Granted, this keeps you from using commas in individual chore
	// descriptions, so the members of your household will have to keep their
	// requests short.
	// As an exercise, figure out a way around this limitation.
	// One approach may be to use a different delimiter, but keep in mind,
	// whatever you choose as a delimiter may have significance on the command
	// line.
	// Another approach may be to add support for quoted strings containing commas.
	for _, chore := range strings.Split(s, ",") {
		if desc := strings.TrimSpace(chore); desc != "" {
			chores = append(chores, &housework.Chore{
				Description: desc,
			})
		}
	}

	// ... and flush the changes to storage.
	return flush(chores)
}

// Page 275
// Listing 12-8: Marking a chore as complete.
// The complete function accepts the command line argument representing the
// chore you want to complete and converts it to an integer.
// You mark only one complete at a time. You then load the chores from storage
// and make sure the integer is within range. If so, you mark the chore
// complete.
func complete(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	chores, err := load()
	if err != nil {
		return err
	}

	if i < 1 || i > len(chores) {
		return fmt.Errorf("chore %d not found", i)
	}

	// Since you're numbering chores starting with 1 when displaying the list,
	// you need to account for placement in the slice by subtracting 1.
	chores[i-1].Complete = true

	// Finally, you flush the chores to storage.
	return flush(chores)
}

// Page 276
// Listing 12-9: The main logic of the housework application.
func main() {
	flag.Parse()

	var err error

	switch strings.ToLower(flag.Arg(0)) {
	case "add":
		err = add(strings.Join(flag.Args()[1:], " "))
	case "complete":
		err = complete(flag.Arg(1))
	}

	if err != nil {
		log.Fatal(err)
	}

	err = list()
	if err != nil {
		log.Fatal(err)
	}
}
