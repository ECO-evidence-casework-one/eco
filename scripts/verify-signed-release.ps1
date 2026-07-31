param(
    [Parameter(Mandatory=$true)]
    [string]$Path
)

$ErrorActionPreference = "Stop"
Get-FileHash $Path -Algorithm SHA256 | Format-List
$sig = Get-AuthenticodeSignature $Path
$sig | Format-List Status, StatusMessage, SignerCertificate, TimeStamperCertificate
if ($sig.Status -ne "Valid") {
    throw "The release does not have a valid trusted Authenticode signature."
}
