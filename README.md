```
// 
s := "47cdf3857f 47cdf3857fbb1f5d93d02835052ed5b3417cf5c2 <Liming Liu> <andliu@microsoft.com> Merged PR 3744385: remove the tunnelgateway because ccp pool is enabled everywhere(ccp chart part)."
s := "406e00c691 406e00c6913d6ccfc93eacc0c69061039bdb57d8 <Seth Goings> <Seth.Goings@microsoft.com> Merged PR 1370426: feat(australiasoutheast): envrcs, acs config --> hcp, and hcp chartsconfig"
compRegEx := regexp.MustCompile(`(?P<first>[0-9a-z]{10}) (?P<second>[0-9a-z]{40}) <(?P<third>[A-Za-z0-9\s]+)> <(?P<fourth>[A-Za-z0-9@.]+)> (?P<fifth>.*)`)
// compRegEx := regexp.MustCompile(`(?P<first>[0-9a-z]{10}) (?P<second>[0-9a-z]{40}) [(?P<third>[a-zA-Z ]+)] [(?P<forth>[A-Za-z0-9@.]+)] [(?P<fifth>Merged*)]`)
m := compRegEx.FindStringSubmatch(s)

for _, v := range m {
    fmt.Println(v)
}

for i, name := range compRegEx.SubexpNames() {
    // if i > 0 && i <= len(match) {
    // 	paramsMap[name] = match[i]
    // }
    fmt.Println(i, name)
}
```