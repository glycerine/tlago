import tlc2.value./*impl.*/IntValue;

// For TLC 2.14 (shipped as part of Toolbox 1.5.8) the correct
// package for IntValue (and all value implementation classes)
// will be tlc2.value.impl.  In other words, the import above
// has to be replaced with:
//import tlc2.value.impl.IntValue;

public class Bits {
    public static IntValue And(IntValue x, IntValue y, IntValue n, IntValue m) {
        return IntValue.gen(x.val & y.val);
    }
}

