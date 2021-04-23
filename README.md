## How to use this wrapper

#### Simple code example

```
whichMetric := "CapRealUSD"
sinceWhen := YesterdaySimpleDate() // "" will start with earliest possible datapoint
UntilWhen := ""                    // most recent datapoint

opts := CMAPIListOptions{Metrics: whichMetric, Start: sinceWhen, End: UntilWhen}
md, err := GetMetricData(context.Background(), *NewCommunityClient(""), &opts)
if err != nil {
    panic(err)
}

res := fmt.Sprintf(
    "latest %s : %s @ %s",
    whichMetric,
    md.Data.Series[0].Values[0],
    md.Data.Series[0].Date,
)

println(res)
// latest CapRealUSD : 357839722701.057414591381506310739636 @ 2021-04-22T00:00:00.000Z
```

#### Public Safety Announcement

At this particular moment, only the community client is implemented. 

This implementation uses V3 of CM API. Unfortunately CM's V4 API introduced a breaking change that 
made parsing the json response quite difficult. 
