FROM gitpod/workspace-full-vnc
RUN sudo apt-get update && \
    sudo apt-get install -y graphviz htop openjdk-11-jre-headless

RUN wget https://github.com/tlaplus/tlaplus/releases/download/v1.8.0/TLAToolbox-1.8.0.deb && \
    sudo dpkg -i TLAToolbox-1.8.0.deb

RUN wget https://github.com/tlaplus/tlaplus/releases/download/v1.8.0/TLAToolbox-1.8.0-linux.gtk.x86_64.zip && \
    unzip TLAToolbox-1.8.0-linux.gtk.x86_64.zip

# You will also need to add "/usr/local/lib/tlaps" to the toolbox's library path 
# (File->Preferences->TLA+ Preferences->Add Directory)
RUN wget https://github.com/tlaplus/tlapm/releases/download/v1.4.5/tlaps-1.4.5-x86_64-linux-gnu-inst.bin && \
    chmod +x tlaps-1.4.5-x86_64-linux-gnu-inst.bin && \
    sudo ./tlaps-1.4.5-x86_64-linux-gnu-inst.bin

RUN echo "PATH=$PATH:/usr/local/bin" >> /home/gitpod/.bashrc