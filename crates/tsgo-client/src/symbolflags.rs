use bitflags::bitflags;

// SymbolFlags definition from TypeScript
// Reference: typescript-go/internal/ast/symbolflags.go
bitflags! {
    /// Symbol flags from TypeScript compiler
    #[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
    pub struct SymbolFlags: u32 {
        /// None
        const NONE = 0;
        /// Variable (var) or parameter
        const FUNCTION_SCOPED_VARIABLE = 1 << 0;
        /// A block-scoped variable (let or const)
        const BLOCK_SCOPED_VARIABLE = 1 << 1;
        /// Property or enum member
        const PROPERTY = 1 << 2;
        /// Enum member
        const ENUM_MEMBER = 1 << 3;
        /// Function
        const FUNCTION = 1 << 4;
        /// Class
        const CLASS = 1 << 5;
        /// Interface
        const INTERFACE = 1 << 6;
        /// Const enum
        const CONST_ENUM = 1 << 7;
        /// Enum
        const REGULAR_ENUM = 1 << 8;
        /// Instantiated module
        const VALUE_MODULE = 1 << 9;
        /// Uninstantiated module
        const NAMESPACE_MODULE = 1 << 10;
        /// Type Literal or mapped type
        const TYPE_LITERAL = 1 << 11;
        /// Object Literal
        const OBJECT_LITERAL = 1 << 12;
        /// Method
        const METHOD = 1 << 13;
        /// Constructor
        const CONSTRUCTOR = 1 << 14;
        /// Get accessor
        const GET_ACCESSOR = 1 << 15;
        /// Set accessor
        const SET_ACCESSOR = 1 << 16;
        /// Call, construct, or index signature
        const SIGNATURE = 1 << 17;
        /// Type parameter
        const TYPE_PARAMETER = 1 << 18;
        /// Type alias
        const TYPE_ALIAS = 1 << 19;
        /// Exported value marker
        const EXPORT_VALUE = 1 << 20;
        /// An alias for another symbol
        const ALIAS = 1 << 21;
        /// Prototype property (no source representation)
        const PROTOTYPE = 1 << 22;
        /// Export * declaration
        const EXPORT_STAR = 1 << 23;
        /// Optional property
        const OPTIONAL = 1 << 24;
        /// Transient symbol (created during type check)
        const TRANSIENT = 1 << 25;
        /// Assignment to property on function acting as declaration
        const ASSIGNMENT = 1 << 26;
        /// Symbol for CommonJS `module` of `module.exports`
        const MODULE_EXPORTS = 1 << 27;
        /// Module contains only const enums or other modules with only const enums
        const CONST_ENUM_ONLY_MODULE = 1 << 28;
        /// Replaceable by method
        const REPLACEABLE_BY_METHOD = 1 << 29;
        /// Flag to signal this is a global lookup
        const GLOBAL_LOOKUP = 1 << 30;

        // Composite flags
        /// Enum = RegularEnum | ConstEnum
        const ENUM = Self::REGULAR_ENUM.bits() | Self::CONST_ENUM.bits();
        /// Variable = FunctionScopedVariable | BlockScopedVariable
        const VARIABLE = Self::FUNCTION_SCOPED_VARIABLE.bits() | Self::BLOCK_SCOPED_VARIABLE.bits();
        /// Value flags
        const VALUE = Self::VARIABLE.bits() | Self::PROPERTY.bits() | Self::ENUM_MEMBER.bits()
            | Self::OBJECT_LITERAL.bits() | Self::FUNCTION.bits() | Self::CLASS.bits()
            | Self::ENUM.bits() | Self::VALUE_MODULE.bits() | Self::METHOD.bits()
            | Self::GET_ACCESSOR.bits() | Self::SET_ACCESSOR.bits();
        /// Type flags
        const TYPE = Self::CLASS.bits() | Self::INTERFACE.bits() | Self::ENUM.bits()
            | Self::ENUM_MEMBER.bits() | Self::TYPE_LITERAL.bits() | Self::TYPE_PARAMETER.bits()
            | Self::TYPE_ALIAS.bits();
        /// Namespace flags
        const NAMESPACE = Self::VALUE_MODULE.bits() | Self::NAMESPACE_MODULE.bits() | Self::ENUM.bits();
        /// Module flags
        const MODULE = Self::VALUE_MODULE.bits() | Self::NAMESPACE_MODULE.bits();
        /// Accessor flags
        const ACCESSOR = Self::GET_ACCESSOR.bits() | Self::SET_ACCESSOR.bits();
        /// Module member flags
        const MODULE_MEMBER = Self::VARIABLE.bits() | Self::FUNCTION.bits() | Self::CLASS.bits()
            | Self::INTERFACE.bits() | Self::ENUM.bits() | Self::MODULE.bits()
            | Self::TYPE_ALIAS.bits() | Self::ALIAS.bits();
        /// Export has local flags
        const EXPORT_HAS_LOCAL = Self::FUNCTION.bits() | Self::CLASS.bits() | Self::ENUM.bits()
            | Self::VALUE_MODULE.bits();
        /// Block scoped flags
        const BLOCK_SCOPED = Self::BLOCK_SCOPED_VARIABLE.bits() | Self::CLASS.bits() | Self::ENUM.bits();
        /// Property or accessor flags
        const PROPERTY_OR_ACCESSOR = Self::PROPERTY.bits() | Self::ACCESSOR.bits();
        /// Class member flags
        const CLASS_MEMBER = Self::METHOD.bits() | Self::ACCESSOR.bits() | Self::PROPERTY.bits();
        /// Export supports default modifier
        const EXPORT_SUPPORTS_DEFAULT_MODIFIER = Self::CLASS.bits() | Self::FUNCTION.bits()
            | Self::INTERFACE.bits();
        /// Classifiable flags
        const CLASSIFIABLE = Self::CLASS.bits() | Self::ENUM.bits() | Self::TYPE_ALIAS.bits()
            | Self::INTERFACE.bits() | Self::TYPE_PARAMETER.bits() | Self::MODULE.bits()
            | Self::ALIAS.bits();
        /// Late binding container flags
        const LATE_BINDING_CONTAINER = Self::CLASS.bits() | Self::INTERFACE.bits()
            | Self::TYPE_LITERAL.bits() | Self::OBJECT_LITERAL.bits() | Self::FUNCTION.bits();
    }
}
