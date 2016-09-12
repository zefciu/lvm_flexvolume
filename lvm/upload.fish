#!/usr/bin/env fish
ssh-add
for mach in (seq 2 4)
    scp lvm vagrant@10.100.0.$mach:
    ssh vagrant@10.100.0.$mach 'sudo mv lvm /usr/libexec/kubernetes/kubelet-plugins/volume/exec/mirantis.com~lvm/lvm'
end
