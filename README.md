# xing-user-enum

This tool will let you extract a contact list by providing a companies name from Xing.
It will save 2 files. One in the format of "patrick.hener" and one in the format of "phener". This tool is meant to be as a basis for password-spray attacks against logins without a second factor authentication.

# Usage

Build it:

```
go build .
```

Use it:

```
./xing-user-enum
```

You have to provide your login credentials and the target company either through the prompt or by providing via environment:

```
XINGUSER=rogerrabbit@acme.corp
XINGTARGET=acme corp
```

It is not recommended to provide password through environment!

The output files will be saved in the current working directory.

# Restrictions

Xing has a limit of 200 API calls after which the tool will not be usable anymore. The limit is bound to your account. At the time of writing I do not know the reset time of this limit.