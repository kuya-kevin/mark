Generating random text: a Markov chain algorithm
Program saves a frequency table to a file, reads a frequency table from a file, and uses the frequencies to determine text output.

Program is runnable with the command: ./mark COMMAND options where COMMAND is either "read" or "generate".

If COMMAND is read then the command is:
./mark read N outfilename infile1 infile2 ....

where outfilename gives the file to save the table to, and infile1 and so on give the files to
read, one after another. The user can specify any number of input files. N specifies the order of the markov chain.
The program reads each input file, creates a frequency table, and then saves that frequency table to the file "outfilename".

If COMMAND is generate then the command is:
./mark generate modelfile n
where modelfile is the name of a file saved using the read command and n is the number of words
to output. Your program should read the frequency table in the file modelfile and then use it to
generate n words of output.
