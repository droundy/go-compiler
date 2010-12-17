Name of project
===============

I don't have a good name for this project.  If you have an idea, let
me know.  For the moment, it's pointless enough that it doesn't
deserve the effort to find a good name.  The current executable is
called `gogo`, but that's only temporary.

A toy go compiler in go
=======================

This project is an experiment in writing a go compiler in go.  I have
no experience in compiler writing or assembly language, so it's mostly
a learning experience for me.

If this project works out, and I have time to keep working on it, I'd
like to use it as a test bed for experimenting with generics
approaches.  In particular, I think that after I've implemented
interfaces myself, I'll have a clearer picture of how generics
could/should work.

Simple approach
===============

I'm using as simple an approach to code generation that I can
imagine.  All function arguments and return values go on the stack.
The result of every expression also goes on the stack.  This means
it's going to be seriously slow.  Optimizations would be possible, but
that isn't (yet) the point of this project.

I also only generate 32-bit x86 code, which is another simplification.
Writing a general-purpose (or even good) compiler is not one of my
goals.  Lots of smart people have worked on that, and done a better
job than I'm likely ever to do.  If I wanted to write a good compiler,
I'd contribute to gccgo or gc, or I'd write one with llvm as a
backend.  My goal, instead, is to write a compiler that I understand
and can readily modify.  And to have fun doing assembly programming.
