template zkOnacci() {
    signal private input nMinusOneFib; 
    signal private input nMinusTwoFib; 

    signal output nFib;

    nFib <== nMinusOneFib + nMinusTwoFib;
}

component main = zkOnacci();