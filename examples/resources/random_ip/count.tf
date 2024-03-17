resource "random_ip" "example" {
  count        = 20
  address_type = "ipv4"
  cidr_range   = "192.168.1.0/28"
}

output "random_ipv4_addresses" {
  value = random_ip.example[*].result
}
