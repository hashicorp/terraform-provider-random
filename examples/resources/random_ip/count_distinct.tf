resource "random_ip" "example" {
  count        = 50
  address_type = "ipv4"
  cidr_range   = "192.168.1.0/28"
}

output "random_distinct_ipv4_addresses" {
  value = distinct(random_ip.example[*].result)
}
