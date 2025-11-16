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


---
app: grep over db


-h
read from stdin, output to stdout
CTRL-C, CTRL-D, SIGPIPE
spinner
color, escape codes
itty
