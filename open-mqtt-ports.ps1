# Store the current execution policy
$originalPolicy = Get-ExecutionPolicy

# Temporarily set execution policy to Bypass for this session
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force

# Open port 1883 (MQTT over TCP, non-secure) for inbound traffic
New-NetFirewallRule -DisplayName "Allow MQTT Inbound 1883" -Direction Inbound -LocalPort 1883 -Protocol TCP -Action Allow

# Open port 8883 (MQTT over SSL/TLS, secure) for inbound traffic
New-NetFirewallRule -DisplayName "Allow MQTT Inbound 8883" -Direction Inbound -LocalPort 8883 -Protocol TCP -Action Allow

# Open port 1883 (MQTT over TCP, non-secure) for outbound traffic
New-NetFirewallRule -DisplayName "Allow MQTT Outbound 1883" -Direction Outbound -LocalPort 1883 -Protocol TCP -Action Allow

# Open port 8883 (MQTT over SSL/TLS, secure) for outbound traffic
New-NetFirewallRule -DisplayName "Allow MQTT Outbound 8883" -Direction Outbound -LocalPort 8883 -Protocol TCP -Action Allow

# Confirmation message
Write-Host "MQTT ports (1883 and 8883) opened successfully for inbound and outbound traffic"

# Restore the original execution policy
Set-ExecutionPolicy -Scope Process -ExecutionPolicy $originalPolicy -Force

# Wait for user to press Enter before exiting
Write-Host "Press Enter to exit..."
[void][System.Console]::ReadLine()
