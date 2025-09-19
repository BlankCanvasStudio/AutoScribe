So, I need to keep a record of whether functions are AI aware or not.

I don't want to spin up a DB, but I want the effeciency of binary search. I'm going to keep 2 
files: `/etc/autoscribe/db/is-ai-aware.txt` and `/etc/autoscribe/db/not-ai-aware.txt`

Each of these files will contain the hash of the `FullName()` of the function, depending on if 
its AI aware or not. The main reasons to do this are: I can determine the number of elements in 
the file given a fixed hash size, a known length & element size means I can use the read ptr as 
an array ptr (meaning I don't need to read the entire file into memory), I can then binary search 
this file quite easily, I (ideally) get space reductions by saving hashes instead of the full 
strings.

At the start of the document function, I need to verify if the program is AI-aware. This means:
    - Hashing FullName()
    - Searching `/etc/autoscribe/db/is-ai-aware.txt`
        - Returning true if found. else continue
    - Searching `/etc/autoscribe/db/not-ai-aware.txt`
        - Returning false if found. else continue
    - Asking AI if it knows the function
        - Get the package details from the FunctionDetails Object
        - hash `FullName()`
        - Save in sorted order in either `is-ai` or `not-ai`, depending on what the result is

