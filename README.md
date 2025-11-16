## Writing Command Line Friendly Applications
+++
title = "Writing Command Line Friendly Applications"
date = "FIXME"
tags = ["golang"]
categories = ["golang"]
url = "FIXME"
author = "mikit"
+++

### Introduction

Go is an excellent choice for writing command line applications.
- It compiles to binary, the client doesn't need to install a runtime to run your app
- It compiles to static executable, the client doesn't need have specific shared libraries to run your app
- It's easy to cross compile to various combination of operating system and architecture
- Go is a joy to write in

However, you need to follow some rules and understand some concepts in order to write good command line applications.
In [The Art of Unix Programming](http://www.catb.org/esr/writings/taoup/html/) Eric Raymond [quotes](http://www.catb.org/esr/writings/taoup/html/ch01s06.html) Doug McIlroy:
> This is the Unix philosophy: Write programs that do one thing and do it well. 
> Write programs to work together. Write programs to handle text streams, because that is a universal interface.

It's a good advice, and we'll see how we can follow it.


As an example, we'll wreite an application that query a logs file.
Logs has the following fields:

- Origin
- Time
- Method
- Path
- StatusCode
- Size

Our code will let the user query by the `Path` field.
We have a `Query` utility function that queries the logs.
I'm not going to show the code, only the function signature since their code isn't the point of this post.


_Note: You can view the full source code at the [GitHub repo](https://github.com/353words/cmdline)_

**Listing 1: Utility Functions**

```go

// Record is a log record.
type Record struct {
	Origin     string
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	Size       int
}

type Filter struct {
	Path string
	// TODO: Other filter fields
}

// Query returns logs from r that match filter.
func Query(r io.Reader, filter Filter) ([]Record, error) {

```

Query gets an `io.Reader` to read the logs from and a filter, it return a slice of matching log records.

### Providing Help

Let's start with simple initial code:

**Listing 2: Initial Code**

```go
095 func main() {
096     pathQuery := os.Args[1]
097     fileName := os.Args[2]
098 
099     file, err := os.Open(fileName)
100     if err != nil {
101         fmt.Fprintf(os.Stderr, "error: %v\n", err)
102         os.Exit(1)
103     }
104     defer file.Close()
105 
106     filter := Filter{
107         Path: pathQuery,
108     }
109 
110     records, err := Query(file, filter)
111     if err != nil {
112         fmt.Fprintf(os.Stderr, "error: query - %v\n", err)
113         os.Exit(1)
114     }
115 
116     for _, r := range records {
117         fmt.Println(r)
118     }
119 }


```

Listing 2 show the initial code.
The `main` function accepts two arguments - a query and a log filename.
On line 99 we open the file, on line 106 we create the filter and on line 110 we call `Query`.
Finally, on lines 116-119 we print out the results.

Say a user gets our app and tries to query it for usage:

**Listing 3: Help Output**

```
$ ./logs --help
panic: runtime error: index out of range [2] with length 2

goroutine 1 [running]:
main.main()
        cmdline/initial/main.go:97 +0x28e
```

Listing 3 shows what happens when the user tries to invoke the built-in help.
Lucky for the user, the program fails.
Image what how angry they will be if they ran `drop-db --help` and get `1,000,000 records deleted`.

The solution is to add support for `--help` (and `-h`).
We're going to use the built-in `flag` package, but [cobra](https://cobra.dev/), [cli](https://cli.urfave.org/) and others also support showing help.

_Note: Using built-in packages reduces a lot of risk that comes with third-party packages._


**Listing 4: Supporting `--help`**

```go
097 func main() {
098     flag.Usage = func() {
099         fmt.Fprintf(os.Stderr, "usage: %s QUERY LOG_FILE\n", path.Base(os.Args[0]))
100         flag.PrintDefaults()
101     }
102     flag.Parse()
103 
104     if flag.NArg() != 2 {
105         fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
106         os.Exit(1)
107     }
108 
109     pathQuery := flag.Arg(0)
110     fileName := flag.Arg(1)
```

Listing 4 shows support for `--help`. On lines 98-101 we create the help messages.
On line 102 we call `flag.Parse` that will invoke `flag.Usage` and exit the program if it sees `--help` or `-h` in the command line.
On line 104 we check that we have the right amount of arguments.
You should **always** validate user input before starting to run your code.
On lines 109 and 110 we use `flag.Arg` instead of `os.Args`, `flag.Arg(i)` returns i'th positional argument from the user.

### Reading from `os.Stdin`

Say logs are compressed, some in `gzip` format and some in `bzip` format.
Adding support to various formats can be a log of work.
Instead, we're going to support reading input from the standard input.
This way, the user can use a command to unpack the log and pipe it to our program.

**Listing 5: Piping Compressed Logs**

```
$ zcat logs.gz| ./logs index71
{piweba3y.prodigy.com 1995-08-01 04:10:22 +0000 UTC GET /shuttle/missions/sts-71/images/index71.gif 200 337360}
{jericho2.microsoft.com 1995-08-01 05:03:35 +0000 UTC GET /shuttle/missions/sts-71/images/index71.gif 200 337360}
{osc_pc3.79.242.202.in-addr.arpa 1995-08-01 06:04:26 +0000 UTC GET /shuttle/missions/sts-71/images/index71.gif 200 337360}
{castles10.castles.com 1995-08-01 06:14:24 +0000 UTC GET /shuttle/missions/sts-71/images/index71.gif 200 73728}
{firewall.dfw.ibm.com 1995-08-01 06:47:51 +0000 UTC GET /shuttle/missions/sts-71/images/index71.gif 200 337360}
...
```

Listing 5 shows how to pipe compressed logs to our program.

The code changes are not that complicated.

**Listing 6: Reading from Standard Input**

```go
097 func main() {
098     flag.Usage = func() {
099         fmt.Fprintf(os.Stderr, "usage: %s QUERY [LOG_FILE]\n", path.Base(os.Args[0]))
100         flag.PrintDefaults()
101     }
102     flag.Parse()
103 
104     if flag.NArg() < 1 || flag.NArg() > 2 {
105         fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
106         os.Exit(1)
107     }
108 
109     var r io.Reader
110     pathQuery := flag.Arg(0)
111     if flag.NArg() == 1 || flag.Arg(1) == "-" {
112         r = os.Stdin
113     } else {
114         fileName := flag.Arg(1)
115 
116         file, err := os.Open(fileName)
117         if err != nil {
118             fmt.Fprintf(os.Stderr, "error: %v\n", err)
119             os.Exit(1)
120         }
121         defer file.Close()
122 
123         r = file
124     }
```

Listing 6 shows how to support read from the standard input.
On line 99 we mark the log file as optional by enclosing it in `[]`.
On line 104 we check that we got 1 or 2 arguments.
On line 111-113 we use `os.Stdin` if the user did not specify the `LOG_FILE` parameter or gave `-` as the log file name.
Finally, on lines 114-123 we open the file as before if we get a file name.

_Note: `-` is the convention for standard input/output in command line applications._

### Cleaning Up

Our program prints to standard output.
This output might be the input to another program, and the receiving program might not consume all the input.
Or, the user might be bored and hit CTRL-C before getting all the output.

In both of these cases, our program will receive a signal.
The your program receive a signal, it's like a panic, and `main` will not run an deferred cleanup code.

_Note: To check this, add a `defer fmt.Println("CLEANUP")` to the code and then run `zcat logs.gz | ./logs i | head`. You will *not* see the `CLEANUP` output.

To make sure cleanup code runs, you need to catch signals.

**Listing 7: Handling Signals**

```go
101 func main() {
102     flag.Usage = func() {
103         fmt.Fprintf(os.Stderr, "usage: %s QUERY [LOG_FILE]\n", path.Base(os.Args[0]))
104         flag.PrintDefaults()
105     }
106     flag.Parse()
107 
108     if flag.NArg() < 1 || flag.NArg() > 2 {
109         fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
110         os.Exit(1)
111     }
112 
113     ch := make(chan os.Signal, 1)
114     signal.Notify(ch, unix.SIGPIPE, unix.SIGINT)
115     go func() {
116         <-ch
117         slog.Info("terminating due to signal")
118         os.Exit(0)
119     }()
```

Listing 7 shows how to handle signals.
On line 113 we create a buffered channel and on line 114 we tell `signal` to send a message to `ch` if the program receives `SIGPIPE` or `SIGTERM`. `SIGPIPE` happens when you print to standard output but it is closed (like in the `head` example above) and `SIGTERM` happens when the user hits `CTRL-C`.
Since signal are platform dependent, we use the `golang.org/x/sys/unix` to get signal values.
On lines 115-119 we create a goroutine that will get notified via `ch` that a signal was received.
The goroutine logs and exits the program.

_Note: Signal can be complicated and are not cross platform, see [the signal package documentation](https://pkg.go.dev/os/signal) and [Wikipedia](https://en.wikipedia.org/wiki/Signal_(IPC)) for more details._

### Colorizing Output

It's hard to see where the term we searched for is in the output.
To help the user, we're going to color the matches. We'll do that by using [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code).
These are special sequences that your terminal interrupts.

**Listing 8: Adding Color**

```go
101 // colorize colors term in red.
102 func colorize(s, term string) string {
103     i := strings.Index(s, term)
104     if i == -1 {
105         return s
106     }
107 
108     return s[:i] + "\033[31m" + term + "\033[39m" + s[i+len(term):]
109 }
...
158     isTTY := isatty.IsTerminal(os.Stdout.Fd())
159 
160     for _, r := range records {
161         s := fmt.Sprintf("%v", r)
162         if isTTY {
163             s = colorize(s, pathQuery)
164         }
165         fmt.Println(s)

```
Listing 8 shows how to add color to the output.
On lines 101-108 we create `colorize` that uses ANSI escape codes to color part of the output in red.
On line 158 we use the `github.com/mattn/go-isatty` to know if we're printing to a terminal (vs writing to a file).
On line 162-164 we add color only if printing to terminal.
And finally, on line 165 we print the output.

_Note: The name `TTY` is an acronym to [Teleprinter](https://en.wikipedia.org/wiki/Teleprinter)._


### Looking Faster

Out code might take its time to run, and we don't want the users to give up.
One of the most common ways is to add progress indicators.
[Studies](https://www.nngroup.com/articles/progress-indicators/) shown that user are willing to wait more time if there's something moving on the screen.

**Listing 9: Adding A Spinner**

```go
154     done := make(chan struct{})
155     isTTY := isatty.IsTerminal(os.Stdout.Fd())
156 
157     if isTTY {
158         go func() {
159             spinners := `-\|/`
160             i := 0
161             ticker := time.NewTicker(100 * time.Millisecond)
162             for {
163                 select {
164                 case <-done:
165                     return
166                 case <-ticker.C:
167                     i = (i + 1) % len(spinners)
168                     fmt.Printf(" %c\r", spinners[i])
169                 }
170             }
171         }()
172     }
173 
174     records, err := Query(r, filter)
175     close(done)
176     if err != nil {
177         fmt.Fprintf(os.Stderr, "error: query - %v\n", err)
178         os.Exit(1)
179     }
180 
181     for _, r := range records {
182         s := fmt.Sprintf("%v", r)
183         if isTTY {
184             s = colorize(s, pathQuery)
185         }
186         fmt.Println(s)
187     }
```
On line 154 we create a `done` channel.
On lines 157-172 we create a goroutine (if in TTY). 
This goroutine creates a ticker and then uses `select` to listen both on the ticker and `done` channel.
On line 167 it increments the index of the current symbol and then prints it with `\r` that will return the cursor to the beginning of the line.

On line 175 we close the `done` channel to stop the spinner and then continue to prints the results as before.

_Note: You can get fancier progress with packages such as `github.com/schollz/progressbar/v3` and `https://github.com/charmbracelet/bubbletea`._

### Summary

Our code got more complicated than the initial version, but now it's a good command line citizen and gives the user a better experience. It's up to you to decide whether this effort is worth it.
In some "one off" program I only read from stdin and write to stdout and that's enough, but once other people want to use my code I add more features to make it user friendly.

The terminal your code runs in has a lot of capabilities, its worth investing the time to learn them.
See [Bubble Tea](https://github.com/charmbracelet/bubbletea) for what you can get to.

But always keep in mind the unix philosophy:

> Write programs that do one thing and do it well. Write programs to work together. Write programs to handle text streams, because that is a universal interface.
