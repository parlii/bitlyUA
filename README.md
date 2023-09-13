### Bitly's Status on ICANN's Universal Acceptance Project

This respositary contains a test script to check if Bitly's APIs supports “Universal Acceptance” of newer, longer and internationalized top-level domains and email addresses.

## Test Result:

| Action   | Description                                                             | Supported |
| -------- | ----------------------------------------------------------------------- | --------- |
| Accept   | URL is shortened                                                        | ✅        |
| Validate | URL is shortened                                                        | ✅        |
| Store    | URL is returned in its original IDN format while retrieving the Bitlink | ✅        |
| Process  | Short URL (Bitlink) redirects to the original IDN URL                   | ✅        |
| Display  | URL is returned in its original IDN format while retrieving the Bitlink | ✅        |

Bitly's API supports Universal Acceptance of newer, longer and internationalized top-level domains addresses.

When Bitly short links are redirected, the IDNA format of the original URL is returned which is then translated by the browser to the correct IDN URL.

## How to run the test script

go test .
