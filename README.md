# Terminal Session Recorder  

This project is a terminal session recorder inspired by [asciinema](https://asciinema.org/). It was designed as a replacement for the default shell, capturing and recording all user session activity.  

> **Note:** This project was written some time ago and is being published as-is. It may serve as a useful reference for anyone tackling similar challenges.  

## Features  
- Captures stdin and pseudoterminal output, saving them to a file or network location  
- Stores records in order by timestamp  
- Saves two types of records:  
  1. **Keystrokes** (commands entered by the user)  
  2. **Terminal Info** (environment variables, terminal size, etc.)  
- File contains metadata about the shell environment  
- Network server:  
  - Stores session recordings  
  - Indexes commands for searching  
  - Generates and serves web-formatted versions of recordings  
- Web server:  
  - Handles HTTP requests  
  - Requests records from the network server  
  - Assists in searching recorded sessions  

## Installation & Usage  
_(Coming soon – additional details on setting up and running the project.)_  

## References  
Here are some resources that were useful during development:  
- [asciinema](https://asciinema.org/)  
- [PTY (Pseudoterminal) Programming](https://man7.org/linux/man-pages/man7/pty.7.html)  
- [Shell and Terminal Basics](https://www.gnu.org/software/libc/manual/html_node/Pseudo_002dTerminal-Pairs.html)  

## License  
GPL v3.0

---

Feel free to customize this further! Do you want to add setup instructions or code examples?
