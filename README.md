# Generating random text: a Markov chain algorithm.

Program saves a frequency table to a file, reads a frequency table from a file, and uses the frequencies to determine text output.

Program is runnable with the command: 
```
./mark COMMAND options
```

where COMMAND is either "read" or "generate".

If COMMAND is "read" then the command is: 
```
./mark read N outfilename infile1 infile2 ....
```

where `outfilename` gives the file to save the table to, and `infile1...` and so on give the files to read. The user can specify any number of input files. `N` specifies the order of the markov chain. The program reads each input file, creates a frequency table, and then saves that frequency table to `outfilename`.

If COMMAND is "generate" then the command is: 
```
./mark generate modelfile N
```
 where `modelfile` is the name of a file saved using the read command and `N` is the number of words to output. The program reads the frequency table in `modelfile` uses it to generate `N` words of output.