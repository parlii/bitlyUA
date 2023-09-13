### Bitly's Status on ICANN's Universal Acceptance Project

This repository contains a test script to check if Bitly's APIs supports “Universal Acceptance” of newer, longer and internationalized top-level domains.

## Test Result:

| Action   | Description                                                             | Supported |
| -------- | ----------------------------------------------------------------------- | --------- |
| Accept   | URL is shortened                                                        | ✅❗️     |
| Validate | URL is shortened                                                        | ✅        |
| Store    | URL is returned in its original IDN format while retrieving the Bitlink | ✅        |
| Process  | Short URL (Bitlink) redirects to the equivalent punycode URL            | ✅ 🔹     |
| Display  | URL is returned in its original IDN format while retrieving the Bitlink | ✅        |

❗️The one exception is URLs containing the Ideographic Full Stop symbol (。), which fails conversion due to limitations in the idna library used by Bitly.

🔹 Bitly short links redirect to the punycode format of the IDN URL which is used for DNS resolution. Browsers are expected to convert and display the punycode URL in the original IDN format.

Bitly's API supports Universal Acceptance of newer, longer and internationalized top-level domains addresses.

## How to run the test script

```
go test .
```
