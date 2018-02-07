#The provided account is "compilers" and the password is "cool".

mkdir examples
cd examples
cp /usr/class/cs143/examples/hello_world.cl .
coolc hello_world.cl
spim hello_world.s

make -f /usr/class/cs143/assignments/PA2J/Makefile

make lexer 
