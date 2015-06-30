# twocents
A stand-alone in-memory weighted prefix autosuggest/autocomplete service

# Usage
## Running 
```
twocents [-d pathToDataFiles] [-p port]
```

The optional pathToDataFiles is the name of a directory containing pipe-delimited txt files of the format
```
string|int
```

For example, in /tmp/dictionaries/defaults.txt:
```
Football Factory|1
Football|94116
Football Association Cup|78
Football|1994
Food Technology Services Inc|1
Food and Drug Administration|5869
Food and Chemical Toxicology|1
Food Cycle Science|1
Food|29960
Food Contamination and Poisoning|3321
Food Stamps|726
Food Additives|348
Food Aid|98
Food Banks and Pantries|73
Food Trucks and Vendors|23
National Football League|8344
National Football League Players Assn|259
Food and Drug Administration|5869
Food Contamination and Poisoning|3321
Football Factory|1
Football|94116
Football Association Cup|78
Football|1994
Fast Food Industry|1627
Food Stamps|726
International Federation of Association Football|575
Snack Foods|363
```

## Requesting
```
curl http://localhost:8080/twocents/v1/<dictionaryName>/<prefixQuery>[/<limit>]
```

For example:
```
curl http://localhost:8080/twocents/v1/default/foo
```

The response contains the phrases with words starting with the query string, in descending order by weight:
```
{
  "suggestions": [
    "Football",
    "Food",
    "National Football League",
    "Food and Drug Administration",
    "Food Contamination and Poisoning",
    "Football",
    "Fast Food Industry",
    "Food Stamps",
    "International Federation of Association Football",
    "Snack Foods"
  ]
}
```

Using the limit parameter:
```
curl http://localhost:8080/twocents/v1/default/foo/6
```

The response contains only the first 6 entries:
```
{
  "suggestions": [
    "Football",
    "Food",
    "National Football League",
    "Food and Drug Administration",
    "Food Contamination and Poisoning",
    "Football"
  ]
}
```

Using the filter parameter:
```
curl http://localhost:8080/twocents/v1/default/foo/1000/fight
```

The response contains only the first 6 entries:
```
{
  "suggestions": [
    "Foo Fighters"
  ]
}
```

