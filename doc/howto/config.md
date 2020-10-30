# Configuration

## JSON config file

JSON config file.


## AWS SDK

The AWS SDK binding self initializes from the environment.
The following environment variables should be set under most circumstances:

* AWS_REGION
* AWS_STS_REGIONAL_ENDPOINTS=regional  

The AWS SDK [documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) includes details on the various ways to configure the SDK.

Getting temporary creds from the admin vm:
```
gen3 arun env | grep AWS | awk '{ print "export " $0 }'
```