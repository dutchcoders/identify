# Identify

[![Build Status - Master](https://travis-ci.org/dutchcoders/identify.svg?branch=master)](https://travis-ci.org/dutchcoders/identify)
[![Project Status](http://opensource.box.com/badges/active.svg)](http://opensource.box.com/badges)
[![Project Status](http://opensource.box.com/badges/maintenance.svg)](http://opensource.box.com/badges)
[![Average time to resolve an issue](http://isitmaintained.com/badge/resolution/dutchcoders/identify.svg)](http://isitmaintained.com/project/dutchcoders/identify "Average time to resolve an issue")
[![Percentage of issues still open](http://isitmaintained.com/badge/open/dutchcoders/identify.svg)](http://isitmaintained.com/project/dutchcoders/identify "Percentage of issues still open")
[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.png?v=103)](https://opensource.org/licenses/GPL-3.0/)

Identify will identify web applications, using a database of file locations and the git repository. While comparing the hashes of the files with the file hashes of the repository it identifies the tag or branch of the web application running.


## Installation

```
go get github.com/dutchcoders/identify
```

## Usage

Parameter | Description | Value
--- | --- | ---
debug | enable debug mode| false
application | application to identify | wordpress, joomla, see db.yaml
no-tags | don't check tags | false
no-branches | don't check branches | false
proxy | use proxy (socks5://127.0.0.1:9050) | none


```
$ identify --application joomla http://joomla.org
Identify - Identify application versions
http://github.com/dutchcoders/identify

DutchSec [https://dutchsec.com/]
========================================
[+] Calculating hashes
[+] Cloning repository
[+] Pulling latest

Web application has been identified as one of the following versions:
- 100% 3.6.3-rc2, 3.6.3, 3.6.4, 3.6.3-rc1, 3.6.3-rc3, 3.6.5
-  75% 3.7.0-alpha1
-  50% 3.7.0-rc1, 3.5.0-beta3, 3.6.0-rc, 3.6.2, 3.7.0-alpha2, 3.7.0-beta4, 3.5.0-rc, 3.6.0, 3.7.0-beta1, 3.5.0-rc2, 3.5.0-rc3, 3.6.1, 3.7.0-beta2, 3.5.0-beta2, 3.5.1, 3.6.1-rc1, staging, 3.5.1-rc2, 3.6.1-rc2, 3.6.0-rc2, 3.7.0-beta3, 3.5.0-rc4, 3.5.1-rc, 3.6.0-alpha, 3.6.0-beta1, 3.6.0-beta2, 3.5.0, 3.5.0-beta4, 3.5.0-beta5
-  25% 3.5.0-beta

$
```

## Disclaimer

Here should come an appropriate disclaimer, no warranties and identify shouldn't be used for malicious intent.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2016 Remco Verhoef (DutchSec).

Code released under [GNU GENERAL PUBLIC LICENSE](LICENSE).
