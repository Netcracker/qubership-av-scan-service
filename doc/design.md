# Design

## Problem

Antivirus Scan Service is required to be able to scan files for viruses.
Scanning should be performed synchronously. 
Files have relatively small size (e.g. ~10Mb).
Antivirus service should provide HTTP API which should synchronously return response telling if file is infected or not.

## Solution

To perform viruses scanning, ClamAV antivirus is used: https://www.clamav.net/. 
Antivirus Scan Service job is to wrap ClamAV in HTTP API. ClamAV itself is a daemon which provides custom TCP API.

ClamAV has some specifics which should be considered:
* Max file size for scanning is 2Gb
* Scanning speed increases linear with file size
* ClamAV requires high RAM (~4Gb), since it loads whole viruses database into memory.
* To keep viruses database updated, it is required to configure supplementary freshclam utility,
    which downloads updated viruses database from provided ClamAV database mirror.

High-level overview of Antivirus Scan Service external behavior looks like following:

![](/doc/images/av-external.png)

High-level overview of Antivirus Scan Service internal structure looks like following:

![](/doc/images/av-internal.png)