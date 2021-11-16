# xing-user-enum

This tool will let you extract a contact list by providing a companies name from Xing.
It will save 2 files. One in the format of "patrick.hener" and one in the format of "phener". This tool is meant to be as a basis for password-spray attacks against logins without a second factor authentication.

# Usage

Simply download executable from [Release Page](https://github.com/patrickhener/xing-user-enum/releases).

Or build it manually:

```
go build .
```

Use it:

```
> ./xing-user-enum
Enter Username: myemailaddress(redacted)
Enter Password:
Enter Target: Thinking Objects
[*] Fetching cookies in www.xing.com
[*] Fetching cookies in login.xing.com
[*] Fetching XSRF Token
[+] XSRF Token found
[*] Trigger login request
[+] Login was successful
[*] There were more than one hits on company
[*] Those are:
[0] Thinking Objects GmbH (https://www.xing.com/pages/thinkingobjectsgmbh)
[1] redacted
[2] redacted
[3] redacted
[4] redacted
[!] Please choose company to use: 0
[+] Target Company is: 'Thinking Objects GmbH'
[*] Query company unique ID
[+] There are '63' employees listed with 'Thinking Objects GmbH' on Xing
[*] Extracting employees in badges of maximum 100
[*] Initial badge
[*] Outputting in format patrick.hener
[+] /home/patrick/tools/recon/xing-user-enum/first.last.users.txt has been written
[*] Outputting in format phener
[+] /home/patrick/tools/recon/xing-user-enum/flast.users.txt has been written
```

You have to provide your login credentials and the target company either through the prompt or by providing via environment:

```
XINGUSER=rogerrabbit@acme.corp
XINGTARGET=acme corp
```

It is not recommended to provide password through environment!

For debugging purposes one could set a proxy like:

```
XINGPROXY=http://127.0.0.1:8080
```

The output files will be saved in the current working directory. Output will be sorted by first name and output will also be all lowercase.