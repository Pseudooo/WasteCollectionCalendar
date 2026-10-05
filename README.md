 
### Overview

In order to view upcoming waste-collection days from my local council you need to go through their website,
enter a postcode, select an address, then view the next immediate collection days or scrub through a clunky
calendar view.

I disliked this, this API hits the gov site to collect calendar days for the next 3 months and returns the
results in the form of an ICS file so it can be added to personal calendars.

### QuickStart

To make use of this you'll need:
1. Your address UPRN _(This is a unique number for an address, you can find this at [findmyaddress.co.uk](https://www.findmyaddress.co.uk/search))_
2. Your postcode

With both of these create a URL:
```
https://bin-days.mitchw.uk/calendar?uprn=<uprn>&postcode=<postcode>
```

This URL can be added to your google calendar _(Or other service)_ as a remote calendar and will add your
waste-collection dates to your calendar as returned by the gov site.

Note: This only works for wiltshire currently

### Run at home

#### Building

This can be built from source if docker is installed on your device. Running the following will build the latest
docker image on your system:
```
docker build .
```

Alternatively an executable can be built directly if you have the go toolchain installed:
```
go build .
```

#### Running

Once running the application will listen on port `8080` and the endpoint can be called.

#### Monitoring

This application is setup for structured json logging and open telemetry for metrics. Logs are exported
to `STDOUT`.

To enable metrics the following environment variables must be set:
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `SERVICE__VERSION`

Otherwise the application will not start.