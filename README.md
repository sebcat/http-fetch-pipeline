I saw something like this in a Rob Pike talk once

## Usage

one fetcher, total time ~= sum of all response times

```
$ time ./http-fetch-pipeline -url-file=foo.txt -n-fetchers=1
http://www.google.se/ 200 OK 132.429848ms
http://www.google.com/ 200 OK 299.752683ms
http://www.sydsvenskan.se/ 200 OK 179.905418ms
http://www.reddit.com/ 200 OK 952.775185ms
0,00s user 0,01s system 0% cpu 1,571 total
```

as many fetchers as URLs, total time ~= slowest response time

```
$ time ./http-fetch-pipeline -url-file=foo.txt -n-fetchers=4 
http://www.sydsvenskan.se/ 200 OK 177.215048ms
http://www.reddit.com/ 200 OK 187.527693ms
http://www.google.se/ 200 OK 195.735808ms
http://www.google.com/ 200 OK 297.999011ms
0,00s user 0,01s system 4% cpu 0,304 total
```


