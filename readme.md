
# AWS Cli Utility

This utility allows for some basic s3 operations in the command line.

    Commands:
    s3-object-ls <bucket> [<search-prefix>]
       Lists the objects in a bucket and optionally searches for a specific prefix
    
    s3-bucket-ls
       Lists buckets
    
    s3-object-dl <bucket> <object-key> <local-file>
       Downloads a specific file to the local file system
    
    s3-bucket-dl <bucket> <local-directory> [<modified-after>]
       Downloads an entire bucket to a local directory.
       <modified-after> should be in the format "2017-12-31"
    
    help
       Prints this information
    
    exit
