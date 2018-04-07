# cartographer

Read `terraform.tfstate` and:
1. Return yml
1. Return yml for outputs with a specific output prefix or no prefix at all
1. Return the map structure of output name to output value

Prefix example:

```
output "jumpbox__tags" {
  value = ["one", "two"]
}

output "director__tags" {
  value = ["three", "four"]
}

output "network" {
  value = "banana"
}
```

Ymlize output:
```
jumpbox__tags:
- one
- two
director__tags:
- three
- four
network: banana
```

YmlizeWithPrefix `director` output:
```
tags:
- three
- four
network: banana
```
