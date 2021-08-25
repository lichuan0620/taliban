# Taliban

An invasive telemetry benchmark tool.

## Build

```bash
docker build -t taliban:latest .
docker run -p 8080:8080 taliban:latest -p examples/histogram.yaml
```

## Use

The service expose metrics generated according to the specified configuration.
Checkout the [examples folder](./examples) for sample configurations. 

```shell
docker run -d -p 8080:8080 taliban:latest
curl localhost:8080/metrics
# HELP naughty_mccarthy Arbitrarily-generated gauge metrics
# TYPE naughty_mccarthy gauge
# naughty_mccarthy{amazing_euler6="sleepy_lamarr4",dazzling_poitras1="vigilant_bell2",great_jemison6="flamboyant_keldysh0"} -1.0110213543693957e+307
# naughty_mccarthy{amazing_euler6="sleepy_lamarr4",dazzling_poitras1="vigilant_bell2",great_jemison6="magical_jackson4"} -8.454367145552644e+307
# naughty_mccarthy{amazing_euler6="sleepy_lamarr4",dazzling_poitras1="vigilant_bell2",great_jemison6="vigorous_rosalind7"} 1.7798864549666633e+308
# naughty_mccarthy{amazing_euler6="sleepy_lamarr4",dazzling_poitras1="vigilant_bell2",great_jemison6="wizardly_khayyam8"} 4.62380796130322e+307
# naughty_mccarthy{amazing_euler6="sleepy_lamarr4",dazzling_poitras1="vigilant_bell2",great_jemison6="zealous_mendel7"} -1.6770941555759105e+307
# ...
```
