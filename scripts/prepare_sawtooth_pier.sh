# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub/
# sawtooth pier-client project we developing.
export PIER_CLIENT_PATH=${WORKSPACE}/pier-client-sawtooth
# bitxhub url that deploy relay-chain
export REMOTE_BITXHUB_PATH=xy@xy:/home/xy/bitxhub/bitxhub
# relay-chain's server ip
export REMOTE_IP="172.19.241.113"

function compile_plugins() {
    cd $PIER_CLIENT_PATH
    make sawtooth
}

function prepare_pier() {
    # delete history config and log
    rm -rf $HOME/.pier
    pier --repo=$HOME/.pier init
    # copy address from relay-chain server and modify pier config file
    scp $REMOTE_BITXHUB_PATH/scripts/build_solo/genesis.json $WORKSPACE
    head -n 18 $HOME/.pier/pier.toml > $HOME/.pier/pier.toml.new
    head -n 6 $WORKSPACE/genesis.json | tail -n -4 >> $HOME/.pier/pier.toml.new
    tail -n 7 $HOME/.pier/pier.toml >> $HOME/.pier/pier.toml.new

    # cat pier.toml.new | sed 
    export TEMP=`head -n 6 $WORKSPACE/genesis.json | tail -n -1`
    echo $TEMP
    cat $HOME/.pier/pier.toml.new | sed '22c \ \ \ \ '"$TEMP"',' > $HOME/.pier/pier.toml
    # exit
    # mv $HOME/.pier/pier.toml.new $HOME/.pier/pier.toml
    sed -i '16c addr = '\"$REMOTE_IP:60011\"'' $HOME/.pier/pier.toml
    sed -i '28c plugin = "sawtooth.so"' $HOME/.pier/pier.toml
    sed -i '29c config = "sawtooth"' $HOME/.pier/pier.toml
    rm $HOME/.pier/pier.toml.new
    # create plugin dir and copy necessary file
    mkdir $HOME/.pier/plugins
    cp $PIER_CLIENT_PATH/build/sawtooth.so $HOME/.pier/plugins/

    mkdir $HOME/.pier/sawtooth
    cp $PIER_CLIENT_PATH/scripts/fabric.validators $HOME/.pier/sawtooth
    cp $PIER_CLIENT_PATH/scripts/fabric_rule.wasm $HOME/.pier/sawtooth

    # register to relay-chain
    pier --repo $HOME/.pier appchain register \
        --name sawtooth \
        --type sawtooth \
        --desc chainD-description \
        --version 1.0.0 \
        --validators $HOME/.pier/sawtooth/fabric.validators

    pier rule deploy --path $HOME/.pier/sawtooth/fabric_rule.wasm
    export PIER_ID=`pier --repo=$HOME/.pier id`
    echo "pier is ready, address is $PIER_ID, you need to save this value for send cross-chain request"
    echo "run following code to start pier"
    echo "pier --repo=$HOME/.pier start"
}
function start() {
    compile_plugins
    prepare_pier
}
start