# !/bin/bash
go build -o build/gestic &&\
chmod +x build/gestic &&\
mkdir -p /home/$USER/.local/bin &&\
cp build/gestic /home/$USER/.local/bin &&\
echo "Installed in ~/.local/bin. Make sure it is in your PATH"
