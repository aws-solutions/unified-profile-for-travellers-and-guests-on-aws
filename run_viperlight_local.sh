 echo "Installing content scanning utility - `pwd`"
wget -q https://viperlight-scanner.s3.amazonaws.com/latest/.viperlightrc
wget -q https://viperlight-scanner.s3.amazonaws.com/latest/codescan-funcs.sh
wget -q https://viperlight-scanner.s3.amazonaws.com/latest/codescan-prebuild-default.sh
wget -q https://viperlight-scanner.s3.amazonaws.com/latest/viperlight.zip
unzip -q viperlight.zip -d ../viperlight
rm -r ./viperlight.zip
echo "Content scanning utility installation complete `date`"
echo "Starting content scanning `date` in `pwd`"
default_scan="./codescan-prebuild-default.sh"
if [ -f "$default_scan" ]; then
    chmod +x $default_scan
else
    default_scan="viperlight scan"
fi
custom_scancmd=./codescan-prebuild-custom.sh

if [ -f "$custom_scancmd" ]; then
    chmod +x $custom_scancmd
    $custom_scancmd
else
    $default_scan
fi
echo "Completed content scanning `date`"
rm .viperlightrc codescan-funcs.sh codescan-prebuild-default.sh
