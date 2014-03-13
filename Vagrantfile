Vagrant::Config.run do |config|
  config.vm.box = "ubuntu1210"
  config.vm.box_url = "http://goo.gl/wxdwM"

  config.vm.forward_port 8080, 8080

  config.vm.provision :shell, :inline => <<-SH
    apt-get update -q
    apt-get install -qy libmagickwand-dev
    apt-get install -qy git
    cd /tmp
    wget -q https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz
    tar xzvf godeb-amd64.tar.gz && rm godeb-amd64.tar.gz
    mv godeb /usr/bin/godeb
    /usr/bin/godeb install
    echo 'export GOPATH=/go' >> /home/vagrant/.bashrc
    GOPATH=/go go get github.com/gographics/imagick/imagick
  SH

  config.vm.share_folder "gopath", "/go", ENV["GOPATH"]
end