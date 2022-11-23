# wordfreq

This tool is made to aggregate, filter and process a set of wordlists.

## Compilation

To compile the project run: `go build -o bin/wordfreq ./main`.

## Usage

The description of `--help` as of 2022-11-23:

```
Usage:
  wordfreq-linux-amd64 [OPTIONS]

Input:
  -w, --wordlist=           The location where we are to consume wordlists from (default: /home/almroot/wordlists)
  -i, --ignore              Ignore filename casing
  -l, --list                List the files that are to be processed
      --include=            Glob pattern of relative file names to include
      --exclude=            Glob pattern of relative file names to exclude (default: .git)

Pre-Processing:
      --pre-include-glob=   Glob pattern of included matches
      --pre-exclude-glob=   Glob pattern of excluded matches
      --pre-include-regex=  Regex pattern of included matches
      --pre-exclude-regex=  Regex pattern of included matches

Processing:
      --order=              The order in which directives are carried out, see supported opcodes below (default: EfaDSfaE)
  -e                        Trims whitespaces before and after the string
  -a=                       Characters to remove from the prefix
  -f=                       Characters to remove from the suffix
  -d=                       Cuts the string based on the delimiter(s) and keep the prefix
  -s=                       Cuts the string based on the delimiter(s) and keep the suffix
  -l                        Normalizes the string to lowercase
  -u                        Normalizes the string to uppercase

Post-Processing:
      --post-include-glob=  Glob pattern of included matches
      --post-exclude-glob=  Glob pattern of excluded matches
      --post-include-regex= Regex pattern of included matches
      --post-exclude-regex= Regex pattern of included matches
      --value-prefix=       A set of strings to prepend to the final string
      --value-suffix=       A set of strings to append to the final string
      --results-max=        The amount of results to return
      --results-freq=       The cut off rate on frequency for which we will abort
      --csv                 Produces a CSV separated by tab, having the frequency and word

Help Options:
  -h, --help                Show this help message
```

### Example

```
almroot@x:/tmp/words$ ls
list1.txt  list2.txt

almroot@x:/tmp/words$ cat list1.txt 
index.php
index.aspx
blog/article.php
.htaccess
/admin/
/console/

almroot@x:/tmp/words$ cat list2.txt 
/api/v1/healthcheck
/api/v2/
/index.php
/index.html
/index.aspx

almroot@x:~(main)$ wordfreq -w /tmp/words/ --csv --include=*.txt -a=/
2       index.php
2       index.aspx
1       .htaccess
1       api/v1/healthcheck
1       blog/article.php
1       admin/
1       console/
1       api/v2/
1       index.html

almroot@x:~(main)$ wordfreq -w /tmp/words/ --csv --include=*.txt -a=/ -d=/
2       api
2       index.php
2       index.aspx
1       index.html
1       blog
1       .htaccess
1       admin
1       console

almroot@x:~(main)$ wordfreq -w /tmp/words/ --csv --include=*.txt -a=/ -d=/ --value-prefix=/
2       /api
2       /index.php
2       /index.aspx
1       /.htaccess
1       /admin
1       /console
1       /index.html
1       /blog

almroot@x:~(main)$ wordfreq -w /tmp/words/ --csv --include=*.txt -a=/ -d=/ --value-prefix=/ --post-include-regex=\\.
2       /index.php
2       /index.aspx
1       /.htaccess
1       /index.html
```