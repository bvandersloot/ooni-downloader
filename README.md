# ooni-downloader
Tool to download bulk data from https://measurements.ooni.torproject.org/api/

## Help message
```
Usage of ooni-downloader:

  ooni-downloader [flags] http_param1:value1 ... http_paramN:valueN

  Information on what HTTP parameters are accepted is available at https://measurements.ooni.torproject.org/api/

  Example: ooni-downloader --output-directory ./output probe_cc:US limit:1000


  -output-directory string
        what folder should contain all of the pulled data files (default "./")
```
