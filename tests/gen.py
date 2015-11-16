from __future__ import print_function

makefile = ""
targets = []

for name in ["ubuntu", "busybox"]:
    rule = """{0}_test.docker: {0}_dockerfile
\tdocker build -t {0}_test -f {0}_dockerfile .
\tdocker save -o {0}_test.docker {0}_test
\tdocker rmi {0}_test
""".format(name)
    targets.append(name + "_test.docker")
    makefile += rule

with open("Makefile", "w+") as f:
    print("build:", " ".join(targets), file=f)
    print(makefile, file=f)
