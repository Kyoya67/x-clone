data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-kernel-6.1-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "nat" {
  ami                         = data.aws_ami.amazon_linux_2023.id
  instance_type               = var.nat_instance_type
  subnet_id                   = var.public_subnet_id
  vpc_security_group_ids      = [var.nat_security_group_id]
  associate_public_ip_address = true
  source_dest_check           = false
  iam_instance_profile        = var.instance_profile_name

  metadata_options {
    http_tokens = "required"
  }

  user_data = <<-EOT
    #!/bin/bash
    set -euxo pipefail

    echo 'net.ipv4.ip_forward = 1' > /etc/sysctl.d/99-nat.conf
    sysctl --system

    dnf install -y iptables-services
    default_interface=$(ip route show default | awk '/default/ {print $5; exit}')
    iptables -t nat -A POSTROUTING -o "$default_interface" -j MASQUERADE
    iptables-save > /etc/sysconfig/iptables
    systemctl enable --now iptables
  EOT

  tags = merge(var.tags, {
    Name = "nat-instance"
    Role = "nat"
  })
}
