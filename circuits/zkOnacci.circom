include "../node_modules/circomlib/circuits/smt/smtverifier.circom";
include "../node_modules/circomlib/circuits/smt/smtprocessor.circom";

/**
 * Process the next number of the fibonacci sequence
 * @param nLevels - merkle tree depth
 * @input senderInput - {Field} - Ethereum address of the sender, used to prevent front running attacks
 * @input stateRoot - {Field} - root of the Merkle tree
 * @input n - {Uint32} - the Nth element of the sequence that is being added
 * @input Fn - {Uint32} - the value of the Nth element of the Fibonacci sequence
 * @input siblingsFn[nLevels] - {Array(Field)} - Siblings Merkle proof to demonstrate that the Nth element of the Fibonacci sequence is NOT already on the tree
 * @input FnMinOne - {Uint32} - the value of the N-1th element of the Fibonacci sequence
 * @input siblingsFnMinOne[nLevels] - {Array(Field)} - Siblings Merkle proof to demonstrate that the N-1th element of the Fibonacci sequence is already on the tree
 * @input FnMinTwo - {Uint32} - the value of the N-2th element of the Fibonacci sequence
 * @input siblingsFnMinTwo[nLevels] - {Array(Field)} - Siblings Merkle proof to demonstrate that the N-2th element of the Fibonacci sequence is already on the tree
 * @output senderOutput - {Field} - address of the sender to avoid front running attacks
 * @output currentRoot - {Field} - root of the Merkle Tree BEFORE adding the next fibonacci element into the tree
 * @output newRoot - {Field} - root of the Merkle Tree After adding the next fibonacci element into the tree
 */
template zkOnacci(nLevels) {
    signal private input senderInput;
    // signal private input stateRoot;
    signal private input n;
    signal private input Fn;
    // signal private input siblingsFn[nLevels];
    signal private input FnMinOne;
    // signal private input siblingsFnMinOne[nLevels];
    signal private input FnMinTwo;
    // signal private input siblingsFnMinTwo[nLevels];

    signal output senderOutput;
    // signal output currentRoot;
    // signal output newRoot;

    var i;

    // TODO: REMOVE!
    n === nLevels;

    // // Proof that n doesn't exist yet
    // // TODO: is this redundant because of the processor?
    // // TODO: is this the correct way to do a non existence proof?
    // component smtFnDoesntExists = SMTVerifier(nLevels);
	// smtFnDoesntExists.enabled <== 1;
	// smtFnDoesntExists.fnc <== 0;
	// smtFnDoesntExists.root <== stateRoot;
	// for (i=0; i<nLevels; i++) {
	// 	smtFnDoesntExists.siblings[i] <== siblingsFn[i];
	// }
	// smtFnDoesntExists.oldKey <== 0;
	// smtFnDoesntExists.oldValue <== 0;
	// smtFnDoesntExists.isOld0 <== 0;
	// smtFnDoesntExists.key <== n;
	// smtFnDoesntExists.value <== 0;

    // // Proof that Fn-1 is already on the tree
    // component smtFnMinOneExists = SMTVerifier(nLevels);
	// smtFnMinOneExists.enabled <== 1;
	// smtFnMinOneExists.fnc <== 0;
	// smtFnMinOneExists.root <== stateRoot;
	// for (i=0; i<nLevels; i++) {
	// 	smtFnMinOneExists.siblings[i] <== siblingsFnMinOne[i];
	// }
	// smtFnMinOneExists.oldKey <== 0;
	// smtFnMinOneExists.oldValue <== 0;
	// smtFnMinOneExists.isOld0 <== 0;
	// smtFnMinOneExists.key <== n-1;
	// smtFnMinOneExists.value <== FnMinOne;
    
    // // Proof that Fn-2 is already on the tree
    // component smtFnMinTwoExists = SMTVerifier(nLevels);
	// smtFnMinTwoExists.enabled <== 1;
	// smtFnMinTwoExists.fnc <== 0;
	// smtFnMinTwoExists.root <== stateRoot;
	// for (i=0; i<nLevels; i++) {
	// 	smtFnMinTwoExists.siblings[i] <== siblingsFnMinTwo[i];
	// }
	// smtFnMinTwoExists.oldKey <== 0;
	// smtFnMinTwoExists.oldValue <== 0;
	// smtFnMinTwoExists.isOld0 <== 0;
	// smtFnMinTwoExists.key <== n-2;
	// smtFnMinTwoExists.value <== FnMinTwo;
    
    // Assert that Fn-2 + Fn-1 = Fn
    Fn === FnMinOne + FnMinTwo
    
    // // Process Fn: add it to the tree to get new root
    // // TODO: Test if this also proofs the non existence of Fn before processing
    // component processor = SMTProcessor(nLevels) ;
    // processor.oldRoot <== stateRoot;
    // for (i = 1; i < nLevels; i++) {
    //     processor.siblings[i] <== siblingsFn[i];
    // }
    // processor.oldKey <== n;
    // processor.oldValue <== 0;
    // processor.isOld0 <== 0;
    // processor.newKey <== n;
    // processor.newValue <== Fn;
    // // TODO: WTF is this?
    // // processor.fnc[0] <== states.P2_fnc0*balanceUpdater.isP2Nop;
    // // processor.fnc[1] <== states.P2_fnc1*balanceUpdater.isP2Nop;

    // Output
    // TODO: should hash the output?
    senderOutput <== senderInput;
    // currentRoot <== stateRoot;
    // newRoot <== processor.newRoot;
}

component main = zkOnacci(5);