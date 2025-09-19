So, I need to keep a record of whether functions have been previously documented

I don't want to spin up a DB, but I want the effeciency of binary search. I'm going to keep 2 
files: `/etc/autoscribe/db/is-documented.txt` and `/etc/autoscribe/db/documentation.txt`

`/etc/autoscribe/db/is-documented.txt` will contain the hash of the `FullName()` of the function, 
and `/etc/autoscribe/db/documentation.txt` will contain the escaped text documentation, with a 
`\n` as the separating character. This way we can easily check if we have documentation, without
needing to worry about filtering through all the documentation text.

To check if a function is documented, we need to verify:
    - That `X.Info.Documentation` is empty
    - That the hash of `FullName()` doesn't exist in `/etc/autoscribe/db/is-documented.txt`
        - If it does exist, we populate `X.Info.Documentation` with the value in `/etc/autoscribe/db/documentation.txt`
        - Else: we use AI to document it

