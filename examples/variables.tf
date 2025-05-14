variable "mac_list" {
  description = "List of MAC addresses with descriptions."
  type = list(object({
    mac_address = string
    description = string
  }))
  default = [
    {
      mac_address = "00:00:00:00:00:19"
      description = "test-1"
    },
    {
      mac_address = "00:00:00:00:00:20"
      description = "test-123"
    },
    {
      mac_address = "00:00:00:00:00:21"
      description = "test-1234"
    },
    {
      mac_address = "00:00:00:00:00:22"
      description = "test-1234"
    }
  ]
}
