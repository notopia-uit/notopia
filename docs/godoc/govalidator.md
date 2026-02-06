## ![](/static/shared/icon/code_gm_grey_24dp.svg) Documentation [¶](#section-documentation "Go to Documentation")

### Overview [¶](#pkg-overview "Go to Overview")

* [Singleton](#hdr-Singleton)
* [Validation Functions Return Type error](#hdr-Validation_Functions_Return_Type_error)
* [Custom Validation Functions](#hdr-Custom_Validation_Functions)
* [Cross-Field Validation](#hdr-Cross_Field_Validation)
* [Multiple Validators](#hdr-Multiple_Validators)
* [Using Validator Tags](#hdr-Using_Validator_Tags)
* [Baked In Validators and Tags](#hdr-Baked_In_Validators_and_Tags)
* [Skip Field](#hdr-Skip_Field)
* [Or Operator](#hdr-Or_Operator)
* [StructOnly](#hdr-StructOnly)
* [NoStructLevel](#hdr-NoStructLevel)
* [Omit Empty](#hdr-Omit_Empty)
* [Omit Nil](#hdr-Omit_Nil)
* [Omit Zero](#hdr-Omit_Zero)
* [Dive](#hdr-Dive)
* [Required](#hdr-Required)
* [Required If](#hdr-Required_If)
* [Required Unless](#hdr-Required_Unless)
* [Required With](#hdr-Required_With)
* [Required With All](#hdr-Required_With_All)
* [Required Without](#hdr-Required_Without)
* [Required Without All](#hdr-Required_Without_All)
* [Excluded If](#hdr-Excluded_If)
* [Excluded Unless](#hdr-Excluded_Unless)
* [Is Default](#hdr-Is_Default)
* [Length](#hdr-Length)
* [Maximum](#hdr-Maximum)
* [Minimum](#hdr-Minimum)
* [Equals](#hdr-Equals)
* [Not Equal](#hdr-Not_Equal)
* [One Of](#hdr-One_Of)
* [One Of Case Insensitive](#hdr-One_Of_Case_Insensitive)
* [Greater Than](#hdr-Greater_Than)
* [Greater Than or Equal](#hdr-Greater_Than_or_Equal)
* [Less Than](#hdr-Less_Than)
* [Less Than or Equal](#hdr-Less_Than_or_Equal)
* [Field Equals Another Field](#hdr-Field_Equals_Another_Field)
* [Field Does Not Equal Another Field](#hdr-Field_Does_Not_Equal_Another_Field)
* [Field Greater Than Another Field](#hdr-Field_Greater_Than_Another_Field)
* [Field Greater Than Another Relative Field](#hdr-Field_Greater_Than_Another_Relative_Field)
* [Field Greater Than or Equal To Another Field](#hdr-Field_Greater_Than_or_Equal_To_Another_Field)
* [Field Greater Than or Equal To Another Relative Field](#hdr-Field_Greater_Than_or_Equal_To_Another_Relative_Field)
* [Less Than Another Field](#hdr-Less_Than_Another_Field)
* [Less Than Another Relative Field](#hdr-Less_Than_Another_Relative_Field)
* [Less Than or Equal To Another Field](#hdr-Less_Than_or_Equal_To_Another_Field)
* [Less Than or Equal To Another Relative Field](#hdr-Less_Than_or_Equal_To_Another_Relative_Field)
* [Field Contains Another Field](#hdr-Field_Contains_Another_Field)
* [Field Excludes Another Field](#hdr-Field_Excludes_Another_Field)
* [Unique](#hdr-Unique)
* [ValidateFn](#hdr-ValidateFn)
* [Alpha Only](#hdr-Alpha_Only)
* [Alpha Space](#hdr-Alpha_Space)
* [Alphanumeric](#hdr-Alphanumeric)
* [Alphanumeric Space](#hdr-Alphanumeric_Space)
* [Alpha Unicode](#hdr-Alpha_Unicode)
* [Alphanumeric Unicode](#hdr-Alphanumeric_Unicode)
* [Boolean](#hdr-Boolean)
* [Number](#hdr-Number)
* [Numeric](#hdr-Numeric)
* [Hexadecimal String](#hdr-Hexadecimal_String)
* [Hexcolor String](#hdr-Hexcolor_String)
* [Lowercase String](#hdr-Lowercase_String)
* [Uppercase String](#hdr-Uppercase_String)
* [RGB String](#hdr-RGB_String)
* [RGBA String](#hdr-RGBA_String)
* [HSL String](#hdr-HSL_String)
* [HSLA String](#hdr-HSLA_String)
* [E.164 Phone Number String](#hdr-E_164_Phone_Number_String)
* [E-mail String](#hdr-E_mail_String)
* [JSON String](#hdr-JSON_String)
* [JWT String](#hdr-JWT_String)
* [File](#hdr-File)
* [Image path](#hdr-Image_path)
* [File Path](#hdr-File_Path)
* [URL String](#hdr-URL_String)
* [URI String](#hdr-URI_String)
* [Urn](#hdr-Urn_RFC_2141_String) [RFC 2141](https://rfc-editor.org/rfc/rfc2141.html) String
* [Base32 String](#hdr-Base32_String)
* [Base64 String](#hdr-Base64_String)
* [Base64URL String](#hdr-Base64URL_String)
* [Base64RawURL String](#hdr-Base64RawURL_String)
* [Bitcoin Address](#hdr-Bitcoin_Address)
* [Ethereum Address](#hdr-Ethereum_Address)
* [Contains](#hdr-Contains)
* [Contains Any](#hdr-Contains_Any)
* [Contains Rune](#hdr-Contains_Rune)
* [Excludes](#hdr-Excludes)
* [Excludes All](#hdr-Excludes_All)
* [Excludes Rune](#hdr-Excludes_Rune)
* [Starts With](#hdr-Starts_With)
* [Ends With](#hdr-Ends_With)
* [Does Not Start With](#hdr-Does_Not_Start_With)
* [Does Not End With](#hdr-Does_Not_End_With)
* [International Standard Book Number](#hdr-International_Standard_Book_Number)
* [International Standard Book Number 10](#hdr-International_Standard_Book_Number_10)
* [International Standard Book Number 13](#hdr-International_Standard_Book_Number_13)
* [Universally Unique Identifier UUID](#hdr-Universally_Unique_Identifier_UUID)
* [Universally Unique Identifier UUID v3](#hdr-Universally_Unique_Identifier_UUID_v3)
* [Universally Unique Identifier UUID v4](#hdr-Universally_Unique_Identifier_UUID_v4)
* [Universally Unique Identifier UUID v5](#hdr-Universally_Unique_Identifier_UUID_v5)
* [Universally Unique Lexicographically Sortable Identifier ULID](#hdr-Universally_Unique_Lexicographically_Sortable_Identifier_ULID)
* [ASCII](#hdr-ASCII)
* [Printable ASCII](#hdr-Printable_ASCII)
* [Multi-Byte Characters](#hdr-Multi_Byte_Characters)
* [Data URL](#hdr-Data_URL)
* [Latitude](#hdr-Latitude)
* [Longitude](#hdr-Longitude)
* [Employeer Identification Number EIN](#hdr-Employeer_Identification_Number_EIN)
* [Social Security Number SSN](#hdr-Social_Security_Number_SSN)
* [Internet Protocol Address IP](#hdr-Internet_Protocol_Address_IP)
* [Internet Protocol Address IPv4](#hdr-Internet_Protocol_Address_IPv4)
* [Internet Protocol Address IPv6](#hdr-Internet_Protocol_Address_IPv6)
* [Classless Inter-Domain Routing CIDR](#hdr-Classless_Inter_Domain_Routing_CIDR)
* [Classless Inter-Domain Routing CIDRv4](#hdr-Classless_Inter_Domain_Routing_CIDRv4)
* [Classless Inter-Domain Routing CIDRv6](#hdr-Classless_Inter_Domain_Routing_CIDRv6)
* [Transmission Control Protocol Address TCP](#hdr-Transmission_Control_Protocol_Address_TCP)
* [Transmission Control Protocol Address TCPv4](#hdr-Transmission_Control_Protocol_Address_TCPv4)
* [Transmission Control Protocol Address TCPv6](#hdr-Transmission_Control_Protocol_Address_TCPv6)
* [User Datagram Protocol Address UDP](#hdr-User_Datagram_Protocol_Address_UDP)
* [User Datagram Protocol Address UDPv4](#hdr-User_Datagram_Protocol_Address_UDPv4)
* [User Datagram Protocol Address UDPv6](#hdr-User_Datagram_Protocol_Address_UDPv6)
* [Internet Protocol Address IP](#hdr-Internet_Protocol_Address_IP-1)
* [Internet Protocol Address IPv4](#hdr-Internet_Protocol_Address_IPv4-1)
* [Internet Protocol Address IPv6](#hdr-Internet_Protocol_Address_IPv6-1)
* [Unix domain socket end point Address](#hdr-Unix_domain_socket_end_point_Address)
* [Unix Domain Socket Exists](#hdr-Unix_Domain_Socket_Exists)
* [Media Access Control Address MAC](#hdr-Media_Access_Control_Address_MAC)
* [Hostname](#hdr-Hostname_RFC_952) [RFC 952](https://rfc-editor.org/rfc/rfc952.html)
* [Hostname](#hdr-Hostname_RFC_1123) [RFC 1123](https://rfc-editor.org/rfc/rfc1123.html)
* [HTML Tags](#hdr-HTML_Tags)
* [HTML Encoded](#hdr-HTML_Encoded)
* [URL Encoded](#hdr-URL_Encoded)
* [Directory](#hdr-Directory)
* [Directory Path](#hdr-Directory_Path)
* [HostPort](#hdr-HostPort)
* [Port](#hdr-Port)
* [Datetime](#hdr-Datetime)
* [Iso3166-1 alpha-2](#hdr-Iso3166_1_alpha_2)
* [Iso3166-1 alpha-3](#hdr-Iso3166_1_alpha_3)
* [Iso3166-1 alpha-numeric](#hdr-Iso3166_1_alpha_numeric)
* [BCP 47 Language Tag](#hdr-BCP_47_Language_Tag)
* [RFC 1035](https://rfc-editor.org/rfc/rfc1035.html) label
* [TimeZone](#hdr-TimeZone)
* [Semantic Version](#hdr-Semantic_Version)
* [CVE Identifier](#hdr-CVE_Identifier)
* [Credit Card](#hdr-Credit_Card)
* [Luhn Checksum](#hdr-Luhn_Checksum)
* [MongoDB](#hdr-MongoDB)
* [Cron](#hdr-Cron)
* [SpiceDb ObjectID/Permission/Object Type](#hdr-SpiceDb_ObjectID_Permission_Object_Type)
* [Alias Validators and Tags](#hdr-Alias_Validators_and_Tags)
* [Non standard validators](#hdr-Non_standard_validators)
* [Panics](#hdr-Panics)

Package validator implements value validations for structs and individual fields
based on tags.

It can also handle Cross-Field and Cross-Struct validation for nested structs
and has the ability to dive into arrays and maps of any type.

see more examples <https://github.com/go-playground/validator/tree/master/_examples>

#### Singleton [¶](#hdr-Singleton "Go to Singleton")

Validator is designed to be thread-safe and used as a singleton instance.
It caches information about your struct and validations,
in essence only parsing your validation tags once per struct type.
Using multiple instances neglects the benefit of caching.
The not thread-safe functions are explicitly marked as such in the documentation.

#### Validation Functions Return Type error [¶](#hdr-Validation_Functions_Return_Type_error "Go to Validation Functions Return Type error")

Doing things this way is actually the way the standard library does, see the
file.Open method here:

```
https://golang.org/pkg/os/#Open.
```

The authors return type "error" to avoid the issue discussed in the following,
where err is always != nil:

```
http://stackoverflow.com/a/29138676/3158232
https://github.com/go-playground/validator/issues/134
```

Validator only InvalidValidationError for bad validation input, nil or
ValidationErrors as type error; so, in your code all you need to do is check
if the error returned is not nil, and if it's not check if error is
InvalidValidationError ( if necessary, most of the time it isn't ) type cast
it to type ValidationErrors like so err.(validator.ValidationErrors).

#### Custom Validation Functions [¶](#hdr-Custom_Validation_Functions "Go to Custom Validation Functions")

Custom Validation functions can be added. Example:

```
// Structure
func customFunc(fl validator.FieldLevel) bool {

	if fl.Field().String() == "invalid" {
		return false
	}

	return true
}

validate.RegisterValidation("custom tag name", customFunc)
// NOTES: using the same tag name as an existing function
//        will overwrite the existing one
```

#### Cross-Field Validation [¶](#hdr-Cross_Field_Validation "Go to Cross-Field Validation")

Cross-Field Validation can be done via the following tags:

* eqfield
* nefield
* gtfield
* gtefield
* ltfield
* ltefield
* eqcsfield
* necsfield
* gtcsfield
* gtecsfield
* ltcsfield
* ltecsfield

If, however, some custom cross-field validation is required, it can be done
using a custom validation.

Why not just have cross-fields validation tags (i.e. only eqcsfield and not
eqfield)?

The reason is efficiency. If you want to check a field within the same struct
"eqfield" only has to find the field on the same struct (1 level). But, if we
used "eqcsfield" it could be multiple levels down. Example:

```
type Inner struct {
	StartDate time.Time
}

type Outer struct {
	InnerStructField *Inner
	CreatedAt time.Time      `validate:"ltecsfield=InnerStructField.StartDate"`
}

now := time.Now()

inner := &Inner{
	StartDate: now,
}

outer := &Outer{
	InnerStructField: inner,
	CreatedAt: now,
}

errs := validate.Struct(outer)

// NOTE: when calling validate.Struct(val) topStruct will be the top level struct passed
//       into the function
//       when calling validate.VarWithValue(val, field, tag) val will be
//       whatever you pass, struct, field...
//       when calling validate.Field(field, tag) val will be nil
```

#### Multiple Validators [¶](#hdr-Multiple_Validators "Go to Multiple Validators")

Multiple validators on a field will process in the order defined. Example:

```
type Test struct {
	Field `validate:"max=10,min=1"`
}

// max will be checked then min
```

Bad Validator definitions are not handled by the library. Example:

```
type Test struct {
	Field `validate:"min=10,max=0"`
}

// this definition of min max will never succeed
```

#### Using Validator Tags [¶](#hdr-Using_Validator_Tags "Go to Using Validator Tags")

Baked In Cross-Field validation only compares fields on the same struct.
If Cross-Field + Cross-Struct validation is needed you should implement your
own custom validator.

Comma (",") is the default separator of validation tags. If you wish to
have a comma included within the parameter (i.e. excludesall=,) you will need to
use the UTF-8 hex representation 0x2C, which is replaced in the code as a comma,
so the above will become excludesall=0x2C.

```
type Test struct {
	Field `validate:"excludesall=,"`    // BAD! Do not include a comma.
	Field `validate:"excludesall=0x2C"` // GOOD! Use the UTF-8 hex representation.
}
```

Pipe ("|") is the 'or' validation tags deparator. If you wish to
have a pipe included within the parameter i.e. excludesall=| you will need to
use the UTF-8 hex representation 0x7C, which is replaced in the code as a pipe,
so the above will become excludesall=0x7C

```
type Test struct {
	Field `validate:"excludesall=|"`    // BAD! Do not include a pipe!
	Field `validate:"excludesall=0x7C"` // GOOD! Use the UTF-8 hex representation.
}
```

#### Baked In Validators and Tags [¶](#hdr-Baked_In_Validators_and_Tags "Go to Baked In Validators and Tags")

Here is a list of the current built in validators:

#### Skip Field [¶](#hdr-Skip_Field "Go to Skip Field")

Tells the validation to skip this struct field; this is particularly
handy in ignoring embedded structs from being validated. (Usage: -)

```
Usage: -
```

#### Or Operator [¶](#hdr-Or_Operator "Go to Or Operator")

This is the 'or' operator allowing multiple validators to be used and
accepted. (Usage: rgb|rgba) <-- this would allow either rgb or rgba
colors to be accepted. This can also be combined with 'and' for example
( Usage: omitempty,rgb|rgba)

```
Usage: |
```

#### StructOnly [¶](#hdr-StructOnly "Go to StructOnly")

When a field that is a nested struct is encountered, and contains this flag
any validation on the nested struct will be run, but none of the nested
struct fields will be validated. This is useful if inside of your program
you know the struct will be valid, but need to verify it has been assigned.
NOTE: only "required" and "omitempty" can be used on a struct itself.

```
Usage: structonly
```

#### NoStructLevel [¶](#hdr-NoStructLevel "Go to NoStructLevel")

Same as structonly tag except that any struct level validations will not run.

```
Usage: nostructlevel
```

#### Omit Empty [¶](#hdr-Omit_Empty "Go to Omit Empty")

Allows conditional validation, for example, if a field is not set with
a value (Determined by the "required" validator) then other validation
such as min or max won't run, but if a value is set validation will run.

```
Usage: omitempty
```

#### Omit Nil [¶](#hdr-Omit_Nil "Go to Omit Nil")

Allows to skip the validation if the value is nil (same as omitempty, but
only for the nil-values).

```
Usage: omitnil
```

#### Omit Zero [¶](#hdr-Omit_Zero "Go to Omit Zero")

Allows to skip the validation if the value is a zero value.
For pointers, it checks if the pointer is nil or the underlying value is a zero value.
For slices and maps, it checks if the value is nil or empty.
Otherwise, behaves the same as omitempty.

```
Usage: omitzero
```

#### Dive [¶](#hdr-Dive "Go to Dive")

This tells the validator to dive into a slice, array or map and validate that
level of the slice, array or map with the validation tags that follow.
Multidimensional nesting is also supported, each level you wish to dive will
require another dive tag. dive has some sub-tags, 'keys' & 'endkeys', please see
the Keys & EndKeys section just below.

```
Usage: dive
```

Example #1

```
[][]string with validation tag "gt=0,dive,len=1,dive,required"
// gt=0 will be applied to []
// len=1 will be applied to []string
// required will be applied to string
```

Example #2

```
[][]string with validation tag "gt=0,dive,dive,required"
// gt=0 will be applied to []
// []string will be spared validation
// required will be applied to string
```

Keys & EndKeys

These are to be used together directly after the dive tag and tells the validator
that anything between 'keys' and 'endkeys' applies to the keys of a map and not the
values; think of it like the 'dive' tag, but for map keys instead of values.
Multidimensional nesting is also supported, each level you wish to validate will
require another 'keys' and 'endkeys' tag. These tags are only valid for maps.

```
Usage: dive,keys,othertagvalidation(s),endkeys,valuevalidationtags
```

Example #1

```
map[string]string with validation tag "gt=0,dive,keys,eq=1|eq=2,endkeys,required"
// gt=0 will be applied to the map itself
// eq=1|eq=2 will be applied to the map keys
// required will be applied to map values
```

Example #2

```
map[[2]string]string with validation tag "gt=0,dive,keys,dive,eq=1|eq=2,endkeys,required"
// gt=0 will be applied to the map itself
// eq=1|eq=2 will be applied to each array element in the map keys
// required will be applied to map values
```

#### Required [¶](#hdr-Required "Go to Required")

This validates that the value is not the data types default zero value.
For numbers ensures value is not zero. For strings ensures value is
not "". For booleans ensures value is not false. For slices, maps, pointers, interfaces, channels and functions
ensures the value is not nil. For structs ensures value is not the zero value when using WithRequiredStructEnabled.

```
Usage: required
```

#### Required If [¶](#hdr-Required_If "Go to Required If")

The field under validation must be present and not empty only if all
the other specified fields are equal to the value following the specified
field. For strings ensures value is not "". For slices, maps, pointers,
interfaces, channels and functions ensures the value is not nil. For structs ensures value is not the zero value.
Using the same field name multiple times in the parameters will result in a panic at runtime.

```
Usage: required_if
```

Examples:

```
// require the field if the Field1 is equal to the parameter given:
Usage: required_if=Field1 foobar

// require the field if the Field1 and Field2 is equal to the value respectively:
Usage: required_if=Field1 foo Field2 bar
```

#### Required Unless [¶](#hdr-Required_Unless "Go to Required Unless")

The field under validation must be present and not empty unless all
the other specified fields are equal to the value following the specified
field. For strings ensures value is not "". For slices, maps, pointers,
interfaces, channels and functions ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: required_unless
```

Examples:

```
// require the field unless the Field1 is equal to the parameter given:
Usage: required_unless=Field1 foobar

// require the field unless the Field1 and Field2 is equal to the value respectively:
Usage: required_unless=Field1 foo Field2 bar
```

#### Required With [¶](#hdr-Required_With "Go to Required With")

The field under validation must be present and not empty only if any
of the other specified fields are present. For strings ensures value is
not "". For slices, maps, pointers, interfaces, channels and functions
ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: required_with
```

Examples:

```
// require the field if the Field1 is present:
Usage: required_with=Field1

// require the field if the Field1 or Field2 is present:
Usage: required_with=Field1 Field2
```

#### Required With All [¶](#hdr-Required_With_All "Go to Required With All")

The field under validation must be present and not empty only if all
of the other specified fields are present. For strings ensures value is
not "". For slices, maps, pointers, interfaces, channels and functions
ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: required_with_all
```

Example:

```
// require the field if the Field1 and Field2 is present:
Usage: required_with_all=Field1 Field2
```

#### Required Without [¶](#hdr-Required_Without "Go to Required Without")

The field under validation must be present and not empty only when any
of the other specified fields are not present. For strings ensures value is
not "". For slices, maps, pointers, interfaces, channels and functions
ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: required_without
```

Examples:

```
// require the field if the Field1 is not present:
Usage: required_without=Field1

// require the field if the Field1 or Field2 is not present:
Usage: required_without=Field1 Field2
```

#### Required Without All [¶](#hdr-Required_Without_All "Go to Required Without All")

The field under validation must be present and not empty only when all
of the other specified fields are not present. For strings ensures value is
not "". For slices, maps, pointers, interfaces, channels and functions
ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: required_without_all
```

Example:

```
// require the field if the Field1 and Field2 is not present:
Usage: required_without_all=Field1 Field2
```

#### Excluded If [¶](#hdr-Excluded_If "Go to Excluded If")

The field under validation must not be present or not empty only if all
the other specified fields are equal to the value following the specified
field. For strings ensures value is not "". For slices, maps, pointers,
interfaces, channels and functions ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: excluded_if
```

Examples:

```
// exclude the field if the Field1 is equal to the parameter given:
Usage: excluded_if=Field1 foobar

// exclude the field if the Field1 and Field2 is equal to the value respectively:
Usage: excluded_if=Field1 foo Field2 bar
```

#### Excluded Unless [¶](#hdr-Excluded_Unless "Go to Excluded Unless")

The field under validation must not be present or empty unless all
the other specified fields are equal to the value following the specified
field. For strings ensures value is not "". For slices, maps, pointers,
interfaces, channels and functions ensures the value is not nil. For structs ensures value is not the zero value.

```
Usage: excluded_unless
```

Examples:

```
// exclude the field unless the Field1 is equal to the parameter given:
Usage: excluded_unless=Field1 foobar

// exclude the field unless the Field1 and Field2 is equal to the value respectively:
Usage: excluded_unless=Field1 foo Field2 bar
```

#### Is Default [¶](#hdr-Is_Default "Go to Is Default")

This validates that the value is the default value and is almost the
opposite of required.

```
Usage: isdefault
```

#### Length [¶](#hdr-Length "Go to Length")

For numbers, length will ensure that the value is
equal to the parameter given. For strings, it checks that
the string length is exactly that number of characters. For slices,
arrays, and maps, validates the number of items.

Example #1

```
Usage: len=10
```

Example #2 (time.Duration)

For time.Duration, len will ensure that the value is equal to the duration given
in the parameter.

```
Usage: len=1h30m
```

#### Maximum [¶](#hdr-Maximum "Go to Maximum")

For numbers, max will ensure that the value is
less than or equal to the parameter given. For strings, it checks
that the string length is at most that number of characters. For
slices, arrays, and maps, validates the number of items.

Example #1

```
Usage: max=10
```

Example #2 (time.Duration)

For time.Duration, max will ensure that the value is less than or equal to the
duration given in the parameter.

```
Usage: max=1h30m
```

#### Minimum [¶](#hdr-Minimum "Go to Minimum")

For numbers, min will ensure that the value is
greater or equal to the parameter given. For strings, it checks that
the string length is at least that number of characters. For slices,
arrays, and maps, validates the number of items.

Example #1

```
Usage: min=10
```

Example #2 (time.Duration)

For time.Duration, min will ensure that the value is greater than or equal to
the duration given in the parameter.

```
Usage: min=1h30m
```

#### Equals [¶](#hdr-Equals "Go to Equals")

For strings & numbers, eq will ensure that the value is
equal to the parameter given. For slices, arrays, and maps,
validates the number of items.

Example #1

```
Usage: eq=10
```

Example #2 (time.Duration)

For time.Duration, eq will ensure that the value is equal to the duration given
in the parameter.

```
Usage: eq=1h30m
```

#### Not Equal [¶](#hdr-Not_Equal "Go to Not Equal")

For strings & numbers, ne will ensure that the value is not
equal to the parameter given. For slices, arrays, and maps,
validates the number of items.

Example #1

```
Usage: ne=10
```

Example #2 (time.Duration)

For time.Duration, ne will ensure that the value is not equal to the duration
given in the parameter.

```
Usage: ne=1h30m
```

#### One Of [¶](#hdr-One_Of "Go to One Of")

For strings, ints, and uints, oneof will ensure that the value
is one of the values in the parameter. The parameter should be
a list of values separated by whitespace. Values may be
strings or numbers. To match strings with spaces in them, include
the target string between single quotes. Kind of like an 'enum'.

```
Usage: oneof=red green
       oneof='red green' 'blue yellow'
       oneof=5 7 9
```

#### One Of Case Insensitive [¶](#hdr-One_Of_Case_Insensitive "Go to One Of Case Insensitive")

Works the same as oneof but is case insensitive and therefore only accepts strings.

```
Usage: oneofci=red green
       oneofci='red green' 'blue yellow'
```

#### Greater Than [¶](#hdr-Greater_Than "Go to Greater Than")

For numbers, this will ensure that the value is greater than the
parameter given. For strings, it checks that the string length
is greater than that number of characters. For slices, arrays
and maps it validates the number of items.

Example #1

```
Usage: gt=10
```

Example #2 (time.Time)

For time.Time ensures the time value is greater than time.Now.UTC().

```
Usage: gt
```

Example #3 (time.Duration)

For time.Duration, gt will ensure that the value is greater than the duration
given in the parameter.

```
Usage: gt=1h30m
```

#### Greater Than or Equal [¶](#hdr-Greater_Than_or_Equal "Go to Greater Than or Equal")

Same as 'min' above. Kept both to make terminology with 'len' easier.

Example #1

```
Usage: gte=10
```

Example #2 (time.Time)

For time.Time ensures the time value is greater than or equal to time.Now.UTC().

```
Usage: gte
```

Example #3 (time.Duration)

For time.Duration, gte will ensure that the value is greater than or equal to
the duration given in the parameter.

```
Usage: gte=1h30m
```

#### Less Than [¶](#hdr-Less_Than "Go to Less Than")

For numbers, this will ensure that the value is less than the parameter given.
For strings, it checks that the string length is less than that number of
characters. For slices, arrays, and maps it validates the number of items.

Example #1

```
Usage: lt=10
```

Example #2 (time.Time)

For time.Time ensures the time value is less than time.Now.UTC().

```
Usage: lt
```

Example #3 (time.Duration)

For time.Duration, lt will ensure that the value is less than the duration given
in the parameter.

```
Usage: lt=1h30m
```

#### Less Than or Equal [¶](#hdr-Less_Than_or_Equal "Go to Less Than or Equal")

Same as 'max' above. Kept both to make terminology with 'len' easier.

Example #1

```
Usage: lte=10
```

Example #2 (time.Time)

For time.Time ensures the time value is less than or equal to time.Now.UTC().

```
Usage: lte
```

Example #3 (time.Duration)

For time.Duration, lte will ensure that the value is less than or equal to the
duration given in the parameter.

```
Usage: lte=1h30m
```

#### Field Equals Another Field [¶](#hdr-Field_Equals_Another_Field "Go to Field Equals Another Field")

This will validate the field value against another fields value either within
a struct or passed in field.

Example #1:

```
// Validation on Password field using:
Usage: eqfield=ConfirmPassword
```

Example #2:

```
// Validating by field:
validate.VarWithValue(password, confirmpassword, "eqfield")
```

Field Equals Another Field (relative)

This does the same as eqfield except that it validates the field provided relative
to the top level struct.

```
Usage: eqcsfield=InnerStructField.Field)
```

#### Field Does Not Equal Another Field [¶](#hdr-Field_Does_Not_Equal_Another_Field "Go to Field Does Not Equal Another Field")

This will validate the field value against another fields value either within
a struct or passed in field.

Examples:

```
// Confirm two colors are not the same:
//
// Validation on Color field:
Usage: nefield=Color2

// Validating by field:
validate.VarWithValue(color1, color2, "nefield")
```

Field Does Not Equal Another Field (relative)

This does the same as nefield except that it validates the field provided
relative to the top level struct.

```
Usage: necsfield=InnerStructField.Field
```

#### Field Greater Than Another Field [¶](#hdr-Field_Greater_Than_Another_Field "Go to Field Greater Than Another Field")

Only valid for Numbers, time.Duration and time.Time types, this will validate
the field value against another fields value either within a struct or passed in
field. usage examples are for validation of a Start and End date:

Example #1:

```
// Validation on End field using:
validate.Struct Usage(gtfield=Start)
```

Example #2:

```
// Validating by field:
validate.VarWithValue(start, end, "gtfield")
```

#### Field Greater Than Another Relative Field [¶](#hdr-Field_Greater_Than_Another_Relative_Field "Go to Field Greater Than Another Relative Field")

This does the same as gtfield except that it validates the field provided
relative to the top level struct.

```
Usage: gtcsfield=InnerStructField.Field
```

#### Field Greater Than or Equal To Another Field [¶](#hdr-Field_Greater_Than_or_Equal_To_Another_Field "Go to Field Greater Than or Equal To Another Field")

Only valid for Numbers, time.Duration and time.Time types, this will validate
the field value against another fields value either within a struct or passed in
field. usage examples are for validation of a Start and End date:

Example #1:

```
// Validation on End field using:
validate.Struct Usage(gtefield=Start)
```

Example #2:

```
// Validating by field:
validate.VarWithValue(start, end, "gtefield")
```

#### Field Greater Than or Equal To Another Relative Field [¶](#hdr-Field_Greater_Than_or_Equal_To_Another_Relative_Field "Go to Field Greater Than or Equal To Another Relative Field")

This does the same as gtefield except that it validates the field provided relative
to the top level struct.

```
Usage: gtecsfield=InnerStructField.Field
```

#### Less Than Another Field [¶](#hdr-Less_Than_Another_Field "Go to Less Than Another Field")

Only valid for Numbers, time.Duration and time.Time types, this will validate
the field value against another fields value either within a struct or passed in
field. usage examples are for validation of a Start and End date:

Example #1:

```
// Validation on End field using:
validate.Struct Usage(ltfield=Start)
```

Example #2:

```
// Validating by field:
validate.VarWithValue(start, end, "ltfield")
```

#### Less Than Another Relative Field [¶](#hdr-Less_Than_Another_Relative_Field "Go to Less Than Another Relative Field")

This does the same as ltfield except that it validates the field provided relative
to the top level struct.

```
Usage: ltcsfield=InnerStructField.Field
```

#### Less Than or Equal To Another Field [¶](#hdr-Less_Than_or_Equal_To_Another_Field "Go to Less Than or Equal To Another Field")

Only valid for Numbers, time.Duration and time.Time types, this will validate
the field value against another fields value either within a struct or passed in
field. usage examples are for validation of a Start and End date:

Example #1:

```
// Validation on End field using:
validate.Struct Usage(ltefield=Start)
```

Example #2:

```
// Validating by field:
validate.VarWithValue(start, end, "ltefield")
```

#### Less Than or Equal To Another Relative Field [¶](#hdr-Less_Than_or_Equal_To_Another_Relative_Field "Go to Less Than or Equal To Another Relative Field")

This does the same as ltefield except that it validates the field provided relative
to the top level struct.

```
Usage: ltecsfield=InnerStructField.Field
```

#### Field Contains Another Field [¶](#hdr-Field_Contains_Another_Field "Go to Field Contains Another Field")

This does the same as contains except for struct fields. It should only be used
with string types. See the behavior of reflect.Value.String() for behavior on
other types.

```
Usage: containsfield=InnerStructField.Field
```

#### Field Excludes Another Field [¶](#hdr-Field_Excludes_Another_Field "Go to Field Excludes Another Field")

This does the same as excludes except for struct fields. It should only be used
with string types. See the behavior of reflect.Value.String() for behavior on
other types.

```
Usage: excludesfield=InnerStructField.Field
```

#### Unique [¶](#hdr-Unique "Go to Unique")

For arrays & slices, unique will ensure that there are no duplicates.
For maps, unique will ensure that there are no duplicate values.
For slices of struct, unique will ensure that there are no duplicate values
in a field of the struct specified via a parameter.

```
// For arrays, slices, and maps:
Usage: unique

// For slices of struct:
Usage: unique=field
```

#### ValidateFn [¶](#hdr-ValidateFn "Go to ValidateFn")

This validates that an object responds to a method that can return error or bool.
By default it expects an interface `Validate() error` and check that the method
does not return an error. Other methods can be specified using two signatures:
If the method returns an error, it check if the return value is nil.
If the method returns a boolean, it checks if the value is true.

```
// to use the default method Validate() error
Usage: validateFn

// to use the custom method IsValid() bool (or error)
Usage: validateFn=IsValid
```

#### Alpha Only [¶](#hdr-Alpha_Only "Go to Alpha Only")

This validates that a string value contains ASCII alpha characters only

```
Usage: alpha
```

#### Alpha Space [¶](#hdr-Alpha_Space "Go to Alpha Space")

This validates that a string value contains ASCII alpha characters and spaces only

```
Usage: alphaspace
```

#### Alphanumeric [¶](#hdr-Alphanumeric "Go to Alphanumeric")

This validates that a string value contains ASCII alphanumeric characters only

```
Usage: alphanum
```

#### Alphanumeric Space [¶](#hdr-Alphanumeric_Space "Go to Alphanumeric Space")

This validates that a string value contains ASCII alphanumeric characters and spaces only

```
Usage: alphanumspace
```

#### Alpha Unicode [¶](#hdr-Alpha_Unicode "Go to Alpha Unicode")

This validates that a string value contains unicode alpha characters only

```
Usage: alphaunicode
```

#### Alphanumeric Unicode [¶](#hdr-Alphanumeric_Unicode "Go to Alphanumeric Unicode")

This validates that a string value contains unicode alphanumeric characters only

```
Usage: alphanumunicode
```

#### Boolean [¶](#hdr-Boolean "Go to Boolean")

This validates that a string value can successfully be parsed into a boolean with strconv.ParseBool

```
Usage: boolean
```

#### Number [¶](#hdr-Number "Go to Number")

This validates that a string value contains number values only.
For integers or float it returns true.

```
Usage: number
```

#### Numeric [¶](#hdr-Numeric "Go to Numeric")

This validates that a string value contains a basic numeric value.
basic excludes exponents etc...
for integers or float it returns true.

```
Usage: numeric
```

#### Hexadecimal String [¶](#hdr-Hexadecimal_String "Go to Hexadecimal String")

This validates that a string value contains a valid hexadecimal.

```
Usage: hexadecimal
```

#### Hexcolor String [¶](#hdr-Hexcolor_String "Go to Hexcolor String")

This validates that a string value contains a valid hex color including
hashtag (#)

```
Usage: hexcolor
```

#### Lowercase String [¶](#hdr-Lowercase_String "Go to Lowercase String")

This validates that a string value contains only lowercase characters. An empty string is not a valid lowercase string.

```
Usage: lowercase
```

#### Uppercase String [¶](#hdr-Uppercase_String "Go to Uppercase String")

This validates that a string value contains only uppercase characters. An empty string is not a valid uppercase string.

```
Usage: uppercase
```

#### RGB String [¶](#hdr-RGB_String "Go to RGB String")

This validates that a string value contains a valid rgb color

```
Usage: rgb
```

#### RGBA String [¶](#hdr-RGBA_String "Go to RGBA String")

This validates that a string value contains a valid rgba color

```
Usage: rgba
```

#### HSL String [¶](#hdr-HSL_String "Go to HSL String")

This validates that a string value contains a valid hsl color

```
Usage: hsl
```

#### HSLA String [¶](#hdr-HSLA_String "Go to HSLA String")

This validates that a string value contains a valid hsla color

```
Usage: hsla
```

#### E.164 Phone Number String [¶](#hdr-E_164_Phone_Number_String "Go to E.164 Phone Number String")

This validates that a string value contains a valid E.164 Phone number
<https://en.wikipedia.org/wiki/E.164> (ex. +1123456789)

```
Usage: e164
```

#### E-mail String [¶](#hdr-E_mail_String "Go to E-mail String")

This validates that a string value contains a valid email
This may not conform to all possibilities of any rfc standard, but neither
does any email provider accept all possibilities.

```
Usage: email
```

#### JSON String [¶](#hdr-JSON_String "Go to JSON String")

This validates that a string value is valid JSON

```
Usage: json
```

#### JWT String [¶](#hdr-JWT_String "Go to JWT String")

This validates that a string value is a valid JWT

```
Usage: jwt
```

#### File [¶](#hdr-File "Go to File")

This validates that a string value contains a valid file path and that
the file exists on the machine.
This is done using os.Stat, which is a platform independent function.

```
Usage: file
```

#### Image path [¶](#hdr-Image_path "Go to Image path")

This validates that a string value contains a valid file path and that
the file exists on the machine and is an image.
This is done using os.Stat and github.com/gabriel-vasile/mimetype

```
Usage: image
```

#### File Path [¶](#hdr-File_Path "Go to File Path")

This validates that a string value contains a valid file path but does not
validate the existence of that file.
This is done using os.Stat, which is a platform independent function.

```
Usage: filepath
```

#### URL String [¶](#hdr-URL_String "Go to URL String")

This validates that a string value contains a valid url
This will accept any url the golang request uri accepts but must contain
a schema for example http:// or rtmp://

```
Usage: url
```

#### URI String [¶](#hdr-URI_String "Go to URI String")

This validates that a string value contains a valid uri
This will accept any uri the golang request uri accepts

```
Usage: uri
```

#### Urn [RFC 2141](https://rfc-editor.org/rfc/rfc2141.html) String [RFC 2141](#hdr-Urn_RFC_2141_String "Go to Urn <a href=") String" aria-label="Go to Urn [RFC 2141](https://rfc-editor.org/rfc/rfc2141.html) String">¶

This validates that a string value contains a valid URN
according to the [RFC 2141](https://rfc-editor.org/rfc/rfc2141.html) spec.

```
Usage: urn_rfc2141
```

#### Base32 String [¶](#hdr-Base32_String "Go to Base32 String")

This validates that a string value contains a valid bas324 value.
Although an empty string is valid base32 this will report an empty string
as an error, if you wish to accept an empty string as valid you can use
this with the omitempty tag.

```
Usage: base32
```

#### Base64 String [¶](#hdr-Base64_String "Go to Base64 String")

This validates that a string value contains a valid base64 value.
Although an empty string is valid base64 this will report an empty string
as an error, if you wish to accept an empty string as valid you can use
this with the omitempty tag.

```
Usage: base64
```

#### Base64URL String [¶](#hdr-Base64URL_String "Go to Base64URL String")

This validates that a string value contains a valid base64 URL safe value
according the RFC4648 spec.
Although an empty string is a valid base64 URL safe value, this will report
an empty string as an error, if you wish to accept an empty string as valid
you can use this with the omitempty tag.

```
Usage: base64url
```

#### Base64RawURL String [¶](#hdr-Base64RawURL_String "Go to Base64RawURL String")

This validates that a string value contains a valid base64 URL safe value,
but without = padding, according the RFC4648 spec, section 3.2.
Although an empty string is a valid base64 URL safe value, this will report
an empty string as an error, if you wish to accept an empty string as valid
you can use this with the omitempty tag.

```
Usage: base64rawurl
```

#### Bitcoin Address [¶](#hdr-Bitcoin_Address "Go to Bitcoin Address")

This validates that a string value contains a valid bitcoin address.
The format of the string is checked to ensure it matches one of the three formats
P2PKH, P2SH and performs checksum validation.

```
Usage: btc_addr
```

Bitcoin Bech32 Address (segwit)

This validates that a string value contains a valid bitcoin Bech32 address as defined
by bip-0173 (<https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki>)
Special thanks to Pieter Wuille for providing reference implementations.

```
Usage: btc_addr_bech32
```

#### Ethereum Address [¶](#hdr-Ethereum_Address "Go to Ethereum Address")

This validates that a string value contains a valid ethereum address.
The format of the string is checked to ensure it matches the standard Ethereum address format.

```
Usage: eth_addr
```

#### Contains [¶](#hdr-Contains "Go to Contains")

This validates that a string value contains the substring value.

```
Usage: contains=@
```

#### Contains Any [¶](#hdr-Contains_Any "Go to Contains Any")

This validates that a string value contains any Unicode code points
in the substring value.

```
Usage: containsany=!@#?
```

#### Contains Rune [¶](#hdr-Contains_Rune "Go to Contains Rune")

This validates that a string value contains the supplied rune value.

```
Usage: containsrune=@
```

#### Excludes [¶](#hdr-Excludes "Go to Excludes")

This validates that a string value does not contain the substring value.

```
Usage: excludes=@
```

#### Excludes All [¶](#hdr-Excludes_All "Go to Excludes All")

This validates that a string value does not contain any Unicode code
points in the substring value.

```
Usage: excludesall=!@#?
```

#### Excludes Rune [¶](#hdr-Excludes_Rune "Go to Excludes Rune")

This validates that a string value does not contain the supplied rune value.

```
Usage: excludesrune=@
```

#### Starts With [¶](#hdr-Starts_With "Go to Starts With")

This validates that a string value starts with the supplied string value

```
Usage: startswith=hello
```

#### Ends With [¶](#hdr-Ends_With "Go to Ends With")

This validates that a string value ends with the supplied string value

```
Usage: endswith=goodbye
```

#### Does Not Start With [¶](#hdr-Does_Not_Start_With "Go to Does Not Start With")

This validates that a string value does not start with the supplied string value

```
Usage: startsnotwith=hello
```

#### Does Not End With [¶](#hdr-Does_Not_End_With "Go to Does Not End With")

This validates that a string value does not end with the supplied string value

```
Usage: endsnotwith=goodbye
```

#### International Standard Book Number [¶](#hdr-International_Standard_Book_Number "Go to International Standard Book Number")

This validates that a string value contains a valid isbn10 or isbn13 value.

```
Usage: isbn
```

#### International Standard Book Number 10 [¶](#hdr-International_Standard_Book_Number_10 "Go to International Standard Book Number 10")

This validates that a string value contains a valid isbn10 value.

```
Usage: isbn10
```

#### International Standard Book Number 13 [¶](#hdr-International_Standard_Book_Number_13 "Go to International Standard Book Number 13")

This validates that a string value contains a valid isbn13 value.

```
Usage: isbn13
```

#### Universally Unique Identifier UUID [¶](#hdr-Universally_Unique_Identifier_UUID "Go to Universally Unique Identifier UUID")

This validates that a string value contains a valid UUID. Uppercase UUID values will not pass - use `uuid\_rfc4122` instead.

```
Usage: uuid
```

#### Universally Unique Identifier UUID v3 [¶](#hdr-Universally_Unique_Identifier_UUID_v3 "Go to Universally Unique Identifier UUID v3")

This validates that a string value contains a valid version 3 UUID. Uppercase UUID values will not pass - use `uuid3\_rfc4122` instead.

```
Usage: uuid3
```

#### Universally Unique Identifier UUID v4 [¶](#hdr-Universally_Unique_Identifier_UUID_v4 "Go to Universally Unique Identifier UUID v4")

This validates that a string value contains a valid version 4 UUID. Uppercase UUID values will not pass - use `uuid4\_rfc4122` instead.

```
Usage: uuid4
```

#### Universally Unique Identifier UUID v5 [¶](#hdr-Universally_Unique_Identifier_UUID_v5 "Go to Universally Unique Identifier UUID v5")

This validates that a string value contains a valid version 5 UUID. Uppercase UUID values will not pass - use `uuid5\_rfc4122` instead.

```
Usage: uuid5
```

#### Universally Unique Lexicographically Sortable Identifier ULID [¶](#hdr-Universally_Unique_Lexicographically_Sortable_Identifier_ULID "Go to Universally Unique Lexicographically Sortable Identifier ULID")

This validates that a string value contains a valid ULID value.

```
Usage: ulid
```

#### ASCII [¶](#hdr-ASCII "Go to ASCII")

This validates that a string value contains only ASCII characters.
NOTE: if the string is blank, this validates as true.

```
Usage: ascii
```

#### Printable ASCII [¶](#hdr-Printable_ASCII "Go to Printable ASCII")

This validates that a string value contains only printable ASCII characters.
NOTE: if the string is blank, this validates as true.

```
Usage: printascii
```

#### Multi-Byte Characters [¶](#hdr-Multi_Byte_Characters "Go to Multi-Byte Characters")

This validates that a string value contains one or more multibyte characters.
NOTE: if the string is blank, this validates as true.

```
Usage: multibyte
```

#### Data URL [¶](#hdr-Data_URL "Go to Data URL")

This validates that a string value contains a valid DataURI.
NOTE: this will also validate that the data portion is valid base64

```
Usage: datauri
```

#### Latitude [¶](#hdr-Latitude "Go to Latitude")

This validates that a string value contains a valid latitude.

```
Usage: latitude
```

#### Longitude [¶](#hdr-Longitude "Go to Longitude")

This validates that a string value contains a valid longitude.

```
Usage: longitude
```

#### Employeer Identification Number EIN [¶](#hdr-Employeer_Identification_Number_EIN "Go to Employeer Identification Number EIN")

This validates that a string value contains a valid U.S. Employer Identification Number.

```
Usage: ein
```

#### Social Security Number SSN [¶](#hdr-Social_Security_Number_SSN "Go to Social Security Number SSN")

This validates that a string value contains a valid U.S. Social Security Number.

```
Usage: ssn
```

#### Internet Protocol Address IP [¶](#hdr-Internet_Protocol_Address_IP "Go to Internet Protocol Address IP")

This validates that a string value contains a valid IP Address.

```
Usage: ip
```

#### Internet Protocol Address IPv4 [¶](#hdr-Internet_Protocol_Address_IPv4 "Go to Internet Protocol Address IPv4")

This validates that a string value contains a valid v4 IP Address.

```
Usage: ipv4
```

#### Internet Protocol Address IPv6 [¶](#hdr-Internet_Protocol_Address_IPv6 "Go to Internet Protocol Address IPv6")

This validates that a string value contains a valid v6 IP Address.

```
Usage: ipv6
```

#### Classless Inter-Domain Routing CIDR [¶](#hdr-Classless_Inter_Domain_Routing_CIDR "Go to Classless Inter-Domain Routing CIDR")

This validates that a string value contains a valid CIDR Address.

```
Usage: cidr
```

#### Classless Inter-Domain Routing CIDRv4 [¶](#hdr-Classless_Inter_Domain_Routing_CIDRv4 "Go to Classless Inter-Domain Routing CIDRv4")

This validates that a string value contains a valid v4 CIDR Address.

```
Usage: cidrv4
```

#### Classless Inter-Domain Routing CIDRv6 [¶](#hdr-Classless_Inter_Domain_Routing_CIDRv6 "Go to Classless Inter-Domain Routing CIDRv6")

This validates that a string value contains a valid v6 CIDR Address.

```
Usage: cidrv6
```

#### Transmission Control Protocol Address TCP [¶](#hdr-Transmission_Control_Protocol_Address_TCP "Go to Transmission Control Protocol Address TCP")

This validates that a string value contains a valid resolvable TCP Address.

```
Usage: tcp_addr
```

#### Transmission Control Protocol Address TCPv4 [¶](#hdr-Transmission_Control_Protocol_Address_TCPv4 "Go to Transmission Control Protocol Address TCPv4")

This validates that a string value contains a valid resolvable v4 TCP Address.

```
Usage: tcp4_addr
```

#### Transmission Control Protocol Address TCPv6 [¶](#hdr-Transmission_Control_Protocol_Address_TCPv6 "Go to Transmission Control Protocol Address TCPv6")

This validates that a string value contains a valid resolvable v6 TCP Address.

```
Usage: tcp6_addr
```

#### User Datagram Protocol Address UDP [¶](#hdr-User_Datagram_Protocol_Address_UDP "Go to User Datagram Protocol Address UDP")

This validates that a string value contains a valid resolvable UDP Address.

```
Usage: udp_addr
```

#### User Datagram Protocol Address UDPv4 [¶](#hdr-User_Datagram_Protocol_Address_UDPv4 "Go to User Datagram Protocol Address UDPv4")

This validates that a string value contains a valid resolvable v4 UDP Address.

```
Usage: udp4_addr
```

#### User Datagram Protocol Address UDPv6 [¶](#hdr-User_Datagram_Protocol_Address_UDPv6 "Go to User Datagram Protocol Address UDPv6")

This validates that a string value contains a valid resolvable v6 UDP Address.

```
Usage: udp6_addr
```

#### Internet Protocol Address IP [¶](#hdr-Internet_Protocol_Address_IP-1 "Go to Internet Protocol Address IP")

This validates that a string value contains a valid resolvable IP Address.

```
Usage: ip_addr
```

#### Internet Protocol Address IPv4 [¶](#hdr-Internet_Protocol_Address_IPv4-1 "Go to Internet Protocol Address IPv4")

This validates that a string value contains a valid resolvable v4 IP Address.

```
Usage: ip4_addr
```

#### Internet Protocol Address IPv6 [¶](#hdr-Internet_Protocol_Address_IPv6-1 "Go to Internet Protocol Address IPv6")

This validates that a string value contains a valid resolvable v6 IP Address.

```
Usage: ip6_addr
```

#### Unix domain socket end point Address [¶](#hdr-Unix_domain_socket_end_point_Address "Go to Unix domain socket end point Address")

This validates that a string value contains a valid Unix Address.

```
Usage: unix_addr
```

#### Unix Domain Socket Exists [¶](#hdr-Unix_Domain_Socket_Exists "Go to Unix Domain Socket Exists")

This validates that a Unix domain socket file exists at the specified path.
It checks both filesystem-based sockets and Linux abstract sockets (prefixed with @).
For filesystem sockets, it verifies the path exists and is a socket file.
For abstract sockets on Linux, it checks /proc/net/unix.

```
Usage: uds_exists
```

#### Media Access Control Address MAC [¶](#hdr-Media_Access_Control_Address_MAC "Go to Media Access Control Address MAC")

This validates that a string value contains a valid MAC Address.

```
Usage: mac
```

Note: See Go's ParseMAC for accepted formats and types:

```
http://golang.org/src/net/mac.go?s=866:918#L29
```

#### Hostname [RFC 952](https://rfc-editor.org/rfc/rfc952.html) [RFC 952](#hdr-Hostname_RFC_952 "Go to Hostname <a href=")" aria-label="Go to Hostname [RFC 952](https://rfc-editor.org/rfc/rfc952.html)">¶

This validates that a string value is a valid Hostname according to [RFC 952](https://rfc-editor.org/rfc/rfc952.html) <https://tools.ietf.org/html/rfc952>

```
Usage: hostname
```

#### Hostname [RFC 1123](https://rfc-editor.org/rfc/rfc1123.html) [RFC 1123](#hdr-Hostname_RFC_1123 "Go to Hostname <a href=")" aria-label="Go to Hostname [RFC 1123](https://rfc-editor.org/rfc/rfc1123.html)">¶

This validates that a string value is a valid Hostname according to [RFC 1123](https://rfc-editor.org/rfc/rfc1123.html) <https://tools.ietf.org/html/rfc1123>

```
Usage: hostname_rfc1123 or if you want to continue to use 'hostname' in your tags, create an alias.
```

Full Qualified Domain Name (FQDN)

This validates that a string value contains a valid FQDN.

```
Usage: fqdn
```

#### HTML Tags [¶](#hdr-HTML_Tags "Go to HTML Tags")

This validates that a string value appears to be an HTML element tag
including those described at <https://developer.mozilla.org/en-US/docs/Web/HTML/Element>

```
Usage: html
```

#### HTML Encoded [¶](#hdr-HTML_Encoded "Go to HTML Encoded")

This validates that a string value is a proper character reference in decimal
or hexadecimal format

```
Usage: html_encoded
```

#### URL Encoded [¶](#hdr-URL_Encoded "Go to URL Encoded")

This validates that a string value is percent-encoded (URL encoded) according
to <https://tools.ietf.org/html/rfc3986#section-2.1>

```
Usage: url_encoded
```

#### Directory [¶](#hdr-Directory "Go to Directory")

This validates that a string value contains a valid directory and that
it exists on the machine.
This is done using os.Stat, which is a platform independent function.

```
Usage: dir
```

#### Directory Path [¶](#hdr-Directory_Path "Go to Directory Path")

This validates that a string value contains a valid directory but does
not validate the existence of that directory.
This is done using os.Stat, which is a platform independent function.
It is safest to suffix the string with os.PathSeparator if the directory
may not exist at the time of validation.

```
Usage: dirpath
```

#### HostPort [¶](#hdr-HostPort "Go to HostPort")

This validates that a string value contains a valid DNS hostname and port that
can be used to validate fields typically passed to sockets and connections.

```
Usage: hostname_port
```

#### Port [¶](#hdr-Port "Go to Port")

This validates that the value falls within the valid port number range of 1 to 65,535.

```
Usage: port
```

#### Datetime [¶](#hdr-Datetime "Go to Datetime")

This validates that a string value is a valid datetime based on the supplied datetime format.
Supplied format must match the official Go time format layout as documented in <https://golang.org/pkg/time/>

```
Usage: datetime=2006-01-02
```

#### Iso3166-1 alpha-2 [¶](#hdr-Iso3166_1_alpha_2 "Go to Iso3166-1 alpha-2")

This validates that a string value is a valid country code based on iso3166-1 alpha-2 standard.
see: <https://www.iso.org/iso-3166-country-codes.html>

```
Usage: iso3166_1_alpha2
```

#### Iso3166-1 alpha-3 [¶](#hdr-Iso3166_1_alpha_3 "Go to Iso3166-1 alpha-3")

This validates that a string value is a valid country code based on iso3166-1 alpha-3 standard.
see: <https://www.iso.org/iso-3166-country-codes.html>

```
Usage: iso3166_1_alpha3
```

#### Iso3166-1 alpha-numeric [¶](#hdr-Iso3166_1_alpha_numeric "Go to Iso3166-1 alpha-numeric")

This validates that a string value is a valid country code based on iso3166-1 alpha-numeric standard.
see: <https://www.iso.org/iso-3166-country-codes.html>

```
Usage: iso3166_1_alpha3
```

#### BCP 47 Language Tag [¶](#hdr-BCP_47_Language_Tag "Go to BCP 47 Language Tag")

This validates that a string value is a valid BCP 47 language tag, as parsed by language.Parse.
More information on <https://pkg.go.dev/golang.org/x/text/language>

```
Usage: bcp47_language_tag
```

BIC (SWIFT code - 2022 standard)

This validates that a string value is a valid Business Identifier Code (SWIFT code), defined in ISO 9362:2022.
More information on <https://www.iso.org/standard/84108.html>

```
Usage: bic
```

BIC (SWIFT code - 2014 standard)

This validates that a string value is a valid Business Identifier Code (SWIFT code), defined in ISO 9362:2014.
More information on <https://www.iso.org/standard/60390.html>

```
Usage: bic_iso_9362_2014
```

#### [RFC 1035](https://rfc-editor.org/rfc/rfc1035.html) label [RFC 1035](#hdr-RFC_1035_label "Go to <a href=") label" aria-label="Go to [RFC 1035](https://rfc-editor.org/rfc/rfc1035.html) label">¶

This validates that a string value is a valid dns [RFC 1035](https://rfc-editor.org/rfc/rfc1035.html) label, defined in [RFC 1035](https://rfc-editor.org/rfc/rfc1035.html).
More information on <https://datatracker.ietf.org/doc/html/rfc1035>

```
Usage: dns_rfc1035_label
```

#### TimeZone [¶](#hdr-TimeZone "Go to TimeZone")

This validates that a string value is a valid time zone based on the time zone database present on the system.
Although empty value and Local value are allowed by time.LoadLocation golang function, they are not allowed by this validator.
More information on <https://golang.org/pkg/time/#LoadLocation>

```
Usage: timezone
```

#### Semantic Version [¶](#hdr-Semantic_Version "Go to Semantic Version")

This validates that a string value is a valid semver version, defined in Semantic Versioning 2.0.0.
More information on <https://semver.org/>

```
Usage: semver
```

#### CVE Identifier [¶](#hdr-CVE_Identifier "Go to CVE Identifier")

This validates that a string value is a valid cve id, defined in cve mitre.
More information on <https://cve.mitre.org/>

```
Usage: cve
```

#### Credit Card [¶](#hdr-Credit_Card "Go to Credit Card")

This validates that a string value contains a valid credit card number using Luhn algorithm.

```
Usage: credit_card
```

#### Luhn Checksum [¶](#hdr-Luhn_Checksum "Go to Luhn Checksum")

```
Usage: luhn_checksum
```

This validates that a string or (u)int value contains a valid checksum using the Luhn algorithm.

#### MongoDB [¶](#hdr-MongoDB "Go to MongoDB")

This validates that a string is a valid 24 character hexadecimal string or valid connection string.

```
Usage: mongodb
       mongodb_connection_string
```

Example:

```
type Test struct {
	ObjectIdField         string `validate:"mongodb"`
	ConnectionStringField string `validate:"mongodb_connection_string"`
}
```

#### Cron [¶](#hdr-Cron "Go to Cron")

This validates that a string value contains a valid cron expression.

```
Usage: cron
```

#### SpiceDb ObjectID/Permission/Object Type [¶](#hdr-SpiceDb_ObjectID_Permission_Object_Type "Go to SpiceDb ObjectID/Permission/Object Type")

This validates that a string is valid for use with SpiceDb for the indicated purpose. If no purpose is given, a purpose of 'id' is assumed.

```
Usage: spicedb=id|permission|type
```

#### Alias Validators and Tags [¶](#hdr-Alias_Validators_and_Tags "Go to Alias Validators and Tags")

Alias Validators and Tags
NOTE: When returning an error, the tag returned in "FieldError" will be
the alias tag unless the dive tag is part of the alias. Everything after the
dive tag is not reported as the alias tag. Also, the "ActualTag" in the before
case will be the actual tag within the alias that failed.

Here is a list of the current built in alias tags:

```
"iscolor"
	alias is "hexcolor|rgb|rgba|hsl|hsla" (Usage: iscolor)
"country_code"
	alias is "iso3166_1_alpha2|iso3166_1_alpha3|iso3166_1_alpha_numeric" (Usage: country_code)
```

Validator notes:

```
regex
	a regex validator won't be added because commas and = signs can be part
	of a regex which conflict with the validation definitions. Although
	workarounds can be made, they take away from using pure regex's.
	Furthermore it's quick and dirty but the regex's become harder to
	maintain and are not reusable, so it's as much a programming philosophy
	as anything.

	In place of this new validator functions should be created; a regex can
	be used within the validator function and even be precompiled for better
	efficiency within regexes.go.

	And the best reason, you can submit a pull request and we can keep on
	adding to the validation library of this package!
```

#### Non standard validators [¶](#hdr-Non_standard_validators "Go to Non standard validators")

A collection of validation rules that are frequently needed but are more
complex than the ones found in the baked in validators.
A non standard validator must be registered manually like you would
with your own custom validation functions.

Example of registration and use:

```
type Test struct {
	TestField string `validate:"yourtag"`
}

t := &Test{
	TestField: "Test"
}

validate := validator.New()
validate.RegisterValidation("yourtag", validators.NotBlank)
```

Here is a list of the current non standard validators:

```
NotBlank
	This validates that the value is not blank or with length zero.
	For strings ensures they do not contain only spaces. For channels, maps, slices and arrays
	ensures they don't have zero length. For others, a non empty value is required.

	Usage: notblank
```

#### Panics [¶](#hdr-Panics "Go to Panics")

This package panics when bad input is provided, this is by design, bad code like
that should not make it to production.

```
type Test struct {
	TestField string `validate:"nonexistentfunction=1"`
}

t := &Test{
	TestField: "Test"
}

validate.Struct(t) // this will panic
```

### Index [¶](#pkg-index "Go to Index")

* [type CustomTypeFunc](#CustomTypeFunc)
* [type FieldError](#FieldError)
* [type FieldLevel](#FieldLevel)
* [type FilterFunc](#FilterFunc)
* [type Func](#Func)
* [type FuncCtx](#FuncCtx)
* [type InvalidValidationError](#InvalidValidationError)
* + [func (e \*InvalidValidationError) Error() string](#InvalidValidationError.Error)
* [type Option](#Option)
* + [func WithPrivateFieldValidation() Option](#WithPrivateFieldValidation)
  + [func WithRequiredStructEnabled() Option](#WithRequiredStructEnabled)
* [type RegisterTranslationsFunc](#RegisterTranslationsFunc)
* [type StructLevel](#StructLevel)
* [type StructLevelFunc](#StructLevelFunc)
* [type StructLevelFuncCtx](#StructLevelFuncCtx)
* [type TagNameFunc](#TagNameFunc)
* [type TranslationFunc](#TranslationFunc)
* [type Validate](#Validate)
* + [func New(options ...Option) \*Validate](#New)
* + [func (v \*Validate) RegisterAlias(alias, tags string)](#Validate.RegisterAlias)
  + [func (v \*Validate) RegisterCustomTypeFunc(fn CustomTypeFunc, types ...interface{})](#Validate.RegisterCustomTypeFunc)
  + [func (v \*Validate) RegisterStructValidation(fn StructLevelFunc, types ...interface{})](#Validate.RegisterStructValidation)
  + [func (v \*Validate) RegisterStructValidationCtx(fn StructLevelFuncCtx, types ...interface{})](#Validate.RegisterStructValidationCtx)
  + [func (v \*Validate) RegisterStructValidationMapRules(rules map[string]string, types ...interface{})](#Validate.RegisterStructValidationMapRules)
  + [func (v \*Validate) RegisterTagNameFunc(fn TagNameFunc)](#Validate.RegisterTagNameFunc)
  + [func (v \*Validate) RegisterTranslation(tag string, trans ut.Translator, registerFn RegisterTranslationsFunc, ...) (err error)](#Validate.RegisterTranslation)
  + [func (v \*Validate) RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error](#Validate.RegisterValidation)
  + [func (v \*Validate) RegisterValidationCtx(tag string, fn FuncCtx, callValidationEvenIfNull ...bool) error](#Validate.RegisterValidationCtx)
  + [func (v \*Validate) SetTagName(name string)](#Validate.SetTagName)
  + [func (v \*Validate) Struct(s interface{}) error](#Validate.Struct)
  + [func (v \*Validate) StructCtx(ctx context.Context, s interface{}) (err error)](#Validate.StructCtx)
  + [func (v \*Validate) StructExcept(s interface{}, fields ...string) error](#Validate.StructExcept)
  + [func (v \*Validate) StructExceptCtx(ctx context.Context, s interface{}, fields ...string) (err error)](#Validate.StructExceptCtx)
  + [func (v \*Validate) StructFiltered(s interface{}, fn FilterFunc) error](#Validate.StructFiltered)
  + [func (v \*Validate) StructFilteredCtx(ctx context.Context, s interface{}, fn FilterFunc) (err error)](#Validate.StructFilteredCtx)
  + [func (v \*Validate) StructPartial(s interface{}, fields ...string) error](#Validate.StructPartial)
  + [func (v \*Validate) StructPartialCtx(ctx context.Context, s interface{}, fields ...string) (err error)](#Validate.StructPartialCtx)
  + [func (v \*Validate) ValidateMap(data map[string]interface{}, rules map[string]interface{}) map[string]interface{}](#Validate.ValidateMap)
  + [func (v Validate) ValidateMapCtx(ctx context.Context, data map[string]interface{}, rules map[string]interface{}) map[string]interface{}](#Validate.ValidateMapCtx)
  + [func (v \*Validate) Var(field interface{}, tag string) error](#Validate.Var)
  + [func (v \*Validate) VarCtx(ctx context.Context, field interface{}, tag string) (err error)](#Validate.VarCtx)
  + [func (v \*Validate) VarWithKey(key string, field interface{}, tag string) error](#Validate.VarWithKey)
  + [func (v \*Validate) VarWithKeyCtx(ctx context.Context, key string, field interface{}, tag string) (err error)](#Validate.VarWithKeyCtx)
  + [func (v \*Validate) VarWithValue(field interface{}, other interface{}, tag string) error](#Validate.VarWithValue)
  + [func (v \*Validate) VarWithValueCtx(ctx context.Context, field interface{}, other interface{}, tag string) (err error)](#Validate.VarWithValueCtx)
* [type ValidationErrors](#ValidationErrors)
* + [func (ve ValidationErrors) Error() string](#ValidationErrors.Error)
  + [func (ve ValidationErrors) Translate(ut ut.Translator) ValidationErrorsTranslations](#ValidationErrors.Translate)
* [type ValidationErrorsTranslations](#ValidationErrorsTranslations)

### Constants [¶](#pkg-constants "Go to Constants")

This section is empty.

### Variables [¶](#pkg-variables "Go to Variables")

This section is empty.

### Functions [¶](#pkg-functions "Go to Functions")

This section is empty.

### Types [¶](#pkg-types "Go to Types")

#### type [CustomTypeFunc](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L72) [¶](#CustomTypeFunc "Go to CustomTypeFunc")

```
type CustomTypeFunc func(field reflect.Value) interface{}
```

CustomTypeFunc allows for overriding or adding custom field type handler functions
field = field value of the type to return a value to be validated
example Valuer from sql drive see <https://golang.org/src/database/sql/driver/types.go?s=1210:1293#L29>

#### type [FieldError](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L76) [¶](#FieldError "Go to FieldError")

```
type FieldError interface {

	// Tag returns the validation tag that failed. if the
	// validation was an alias, this will return the
	// alias name and not the underlying tag that failed.
	//
	// eg. alias "iscolor": "hexcolor|rgb|rgba|hsl|hsla"
	// will return "iscolor"
	Tag() string

	// ActualTag returns the validation tag that failed, even if an
	// alias the actual tag within the alias will be returned.
	// If an 'or' validation fails the entire or will be returned.
	//
	// eg. alias "iscolor": "hexcolor|rgb|rgba|hsl|hsla"
	// will return "hexcolor|rgb|rgba|hsl|hsla"
	ActualTag() string

	// Namespace returns the namespace for the field error, with the tag
	// name taking precedence over the field's actual name.
	//
	// eg. JSON name "User.fname"
	//
	// See StructNamespace() for a version that returns actual names.
	//
	// NOTE: this field can be blank when validating a single primitive field
	// using validate.Field(...) as there is no way to extract it's name
	Namespace() string

	// StructNamespace returns the namespace for the field error, with the field's
	// actual name.
	//
	// eg. "User.FirstName" see Namespace for comparison
	//
	// NOTE: this field can be blank when validating a single primitive field
	// using validate.Field(...) as there is no way to extract its name
	StructNamespace() string

	// Field returns the field's name with the tag name taking precedence over the
	// field's actual name.
	//
	// `RegisterTagNameFunc` must be registered to get tag value.
	//
	// eg. JSON name "fname"
	// see StructField for comparison
	Field() string

	// StructField returns the field's actual name from the struct, when able to determine.
	//
	// eg.  "FirstName"
	// see Field for comparison
	StructField() string

	// Value returns the actual field's value in case needed for creating the error
	// message
	Value() interface{}

	// Param returns the param value, in string form for comparison; this will also
	// help with generating an error message
	Param() string

	// Kind returns the Field's reflect Kind
	//
	// eg. time.Time's kind is a struct
	Kind() reflect.Kind

	// Type returns the Field's reflect Type
	//
	// eg. time.Time's type is time.Time
	Type() reflect.Type

	// Translate returns the FieldError's translated error
	// from the provided 'ut.Translator' and registered 'TranslationFunc'
	//
	// NOTE: if no registered translator can be found it returns the same as
	// calling fe.Error()
	Translate(ut ut.Translator) string

	// Error returns the FieldError's message
	Error() string
}
```

FieldError contains all functions to get error details

#### type [FieldLevel](https://github.com/go-playground/validator/blob/v10.30.1/field_level.go#L7) [¶](#FieldLevel "Go to FieldLevel")

```
type FieldLevel interface {

	// Top returns the top level struct, if any
	Top() reflect.Value

	// Parent returns the current fields parent struct, if any or
	// the comparison value if called 'VarWithValue'
	Parent() reflect.Value

	// Field returns current field for validation
	Field() reflect.Value

	// FieldName returns the field's name with the tag
	// name taking precedence over the fields actual name.
	FieldName() string

	// StructFieldName returns the struct field's name
	StructFieldName() string

	// Param returns param for validation against current field
	Param() string

	// GetTag returns the current validations tag name
	GetTag() string

	// ExtractType gets the actual underlying type of field value.
	// It will dive into pointers, customTypes and return you the
	// underlying value and it's kind.
	ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool)

	// GetStructFieldOK traverses the parent struct to retrieve a specific field denoted by the provided namespace
	// in the param and returns the field, field kind and whether is was successful in retrieving
	// the field at all.
	//
	// NOTE: when not successful ok will be false, this can happen when a nested struct is nil and so the field
	// could not be retrieved because it didn't exist.
	//
	// Deprecated: Use GetStructFieldOK2() instead which also return if the value is nullable.
	GetStructFieldOK() (reflect.Value, reflect.Kind, bool)

	// GetStructFieldOKAdvanced is the same as GetStructFieldOK except that it accepts the parent struct to start looking for
	// the field and namespace allowing more extensibility for validators.
	//
	// Deprecated: Use GetStructFieldOKAdvanced2() instead which also return if the value is nullable.
	GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool)

	// GetStructFieldOK2 traverses the parent struct to retrieve a specific field denoted by the provided namespace
	// in the param and returns the field, field kind, if it's a nullable type and whether is was successful in retrieving
	// the field at all.
	//
	// NOTE: when not successful ok will be false, this can happen when a nested struct is nil and so the field
	// could not be retrieved because it didn't exist.
	GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool)

	// GetStructFieldOKAdvanced2 is the same as GetStructFieldOK except that it accepts the parent struct to start looking for
	// the field and namespace allowing more extensibility for validators.
	GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool)
}
```

FieldLevel contains all the information and helper functions
to validate a field

#### type [FilterFunc](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L67) [¶](#FilterFunc "Go to FilterFunc")

```
type FilterFunc func(ns []byte) bool
```

FilterFunc is the type used to filter fields using
StructFiltered(...) function.
returning true results in the field being filtered/skipped from
validation

#### type [Func](https://github.com/go-playground/validator/blob/v10.30.1/baked_in.go#L36) [¶](#Func "Go to Func")

```
type Func func(fl FieldLevel) bool
```

Func accepts a FieldLevel interface for all validation needs. The return
value should be true when validation succeeds.

#### type [FuncCtx](https://github.com/go-playground/validator/blob/v10.30.1/baked_in.go#L40) [¶](#FuncCtx "Go to FuncCtx")

```
type FuncCtx func(ctx context.Context, fl FieldLevel) bool
```

FuncCtx accepts a context.Context and FieldLevel interface for all
validation needs. The return value should be true when validation succeeds.

#### type [InvalidValidationError](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L21) [¶](#InvalidValidationError "Go to InvalidValidationError")

```
type InvalidValidationError struct {
	Type reflect.Type
}
```

InvalidValidationError describes an invalid argument passed to
`Struct`, `StructExcept`, StructPartial` or `Field`

#### func (\*InvalidValidationError) [Error](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L26) [¶](#InvalidValidationError.Error "Go to InvalidValidationError.Error")

```
func (e *InvalidValidationError) Error() string
```

Error returns InvalidValidationError message

#### type [Option](https://github.com/go-playground/validator/blob/v10.30.1/options.go#L4) [¶](#Option "Go to Option") added in v10.15.2

```
type Option func(*Validate)
```

Option represents a configurations option to be applied to validator during initialization.

#### func [WithPrivateFieldValidation](https://github.com/go-playground/validator/blob/v10.30.1/options.go#L22) [¶](#WithPrivateFieldValidation "Go to WithPrivateFieldValidation") added in v10.19.0

```
func WithPrivateFieldValidation() Option
```

WithPrivateFieldValidation activates validation for unexported fields via the use of the `unsafe` package.

By opting into this feature you are acknowledging that you are aware of the risks and accept any current or future
consequences of using this feature.

#### func [WithRequiredStructEnabled](https://github.com/go-playground/validator/blob/v10.30.1/options.go#L12) [¶](#WithRequiredStructEnabled "Go to WithRequiredStructEnabled") added in v10.15.2

```
func WithRequiredStructEnabled() Option
```

WithRequiredStructEnabled enables required tag on non-pointer structs to be applied instead of ignored.

This was made opt-in behaviour in order to maintain backward compatibility with the behaviour previous
to being able to apply struct level validations on struct fields directly.

It is recommended you enabled this as it will be the default behaviour in v11+

#### type [RegisterTranslationsFunc](https://github.com/go-playground/validator/blob/v10.30.1/translations.go#L11) [¶](#RegisterTranslationsFunc "Go to RegisterTranslationsFunc")

```
type RegisterTranslationsFunc func(ut ut.Translator) error
```

RegisterTranslationsFunc allows for registering of translations
for a 'ut.Translator' for use within the 'TranslationFunc'

#### type [StructLevel](https://github.com/go-playground/validator/blob/v10.30.1/struct_level.go#L24) [¶](#StructLevel "Go to StructLevel")

```
type StructLevel interface {

	// Validator returns the main validation object, in case one wants to call validations internally.
	// this is so you don't have to use anonymous functions to get access to the validate
	// instance.
	Validator() *Validate

	// Top returns the top level struct, if any
	Top() reflect.Value

	// Parent returns the current fields parent struct, if any
	Parent() reflect.Value

	// Current returns the current struct.
	Current() reflect.Value

	// ExtractType gets the actual underlying type of field value.
	// It will dive into pointers, customTypes and return you the
	// underlying value and its kind.
	ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool)

	// ReportError reports an error just by passing the field and tag information
	//
	// NOTES:
	//
	// fieldName and structFieldName get appended to the existing
	// namespace that validator is on. e.g. pass 'FirstName' or
	// 'Names[0]' depending on the nesting
	//
	// tag can be an existing validation tag or just something you make up
	// and process on the flip side it's up to you.
	ReportError(field interface{}, fieldName, structFieldName string, tag, param string)

	// ReportValidationErrors reports an error just by passing ValidationErrors
	//
	// NOTES:
	//
	// relativeNamespace and relativeActualNamespace get appended to the
	// existing namespace that validator is on.
	// e.g. pass 'User.FirstName' or 'Users[0].FirstName' depending
	// on the nesting. most of the time they will be blank, unless you validate
	// at a level lower the current field depth
	ReportValidationErrors(relativeNamespace, relativeActualNamespace string, errs ValidationErrors)
}
```

StructLevel contains all the information and helper functions
to validate a struct

#### type [StructLevelFunc](https://github.com/go-playground/validator/blob/v10.30.1/struct_level.go#L9) [¶](#StructLevelFunc "Go to StructLevelFunc")

```
type StructLevelFunc func(sl StructLevel)
```

StructLevelFunc accepts all values needed for struct level validation

#### type [StructLevelFuncCtx](https://github.com/go-playground/validator/blob/v10.30.1/struct_level.go#L13) [¶](#StructLevelFuncCtx "Go to StructLevelFuncCtx")

```
type StructLevelFuncCtx func(ctx context.Context, sl StructLevel)
```

StructLevelFuncCtx accepts all values needed for struct level validation
but also allows passing of contextual validation information via context.Context.

#### type [TagNameFunc](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L75) [¶](#TagNameFunc "Go to TagNameFunc")

```
type TagNameFunc func(field reflect.StructField) string
```

TagNameFunc allows for adding of a custom tag name parser

#### type [TranslationFunc](https://github.com/go-playground/validator/blob/v10.30.1/translations.go#L7) [¶](#TranslationFunc "Go to TranslationFunc")

```
type TranslationFunc func(ut ut.Translator, fe FieldError) string
```

TranslationFunc is the function type used to register or override
custom translations

#### type [Validate](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L83) [¶](#Validate "Go to Validate")

```
type Validate struct {
	// contains filtered or unexported fields
}
```

Validate contains the validator settings and cache

#### func [New](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L106) [¶](#New "Go to New")

```
func New(options ...Option) *Validate
```

New returns a new instance of 'validate' with sane defaults.
Validate is designed to be thread-safe and used as a singleton instance.
It caches information about your struct and validations,
in essence only parsing your validation tags once per struct type.
Using multiple instances neglects the benefit of caching.

#### func (\*Validate) [RegisterAlias](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L239) [¶](#Validate.RegisterAlias "Go to Validate.RegisterAlias")

```
func (v *Validate) RegisterAlias(alias, tags string)
```

RegisterAlias registers a mapping of a single validation tag that
defines a common or complex set of validation(s) to simplify adding validation
to structs.

NOTE: this function is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterCustomTypeFunc](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L308) [¶](#Validate.RegisterCustomTypeFunc "Go to Validate.RegisterCustomTypeFunc")

```
func (v *Validate) RegisterCustomTypeFunc(fn CustomTypeFunc, types ...interface{})
```

RegisterCustomTypeFunc registers a CustomTypeFunc against a number of types

NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterStructValidation](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L253) [¶](#Validate.RegisterStructValidation "Go to Validate.RegisterStructValidation")

```
func (v *Validate) RegisterStructValidation(fn StructLevelFunc, types ...interface{})
```

RegisterStructValidation registers a StructLevelFunc against a number of types.

NOTE:
- this method is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterStructValidationCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L262) [¶](#Validate.RegisterStructValidationCtx "Go to Validate.RegisterStructValidationCtx")

```
func (v *Validate) RegisterStructValidationCtx(fn StructLevelFuncCtx, types ...interface{})
```

RegisterStructValidationCtx registers a StructLevelFuncCtx against a number of types and allows passing
of contextual validation information via context.Context.

NOTE:
- this method is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterStructValidationMapRules](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L281) [¶](#Validate.RegisterStructValidationMapRules "Go to Validate.RegisterStructValidationMapRules") added in v10.11.0

```
func (v *Validate) RegisterStructValidationMapRules(rules map[string]string, types ...interface{})
```

RegisterStructValidationMapRules registers validate map rules.
Be aware that map validation rules supersede those defined on a/the struct if present.

NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterTagNameFunc](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L210) [¶](#Validate.RegisterTagNameFunc "Go to Validate.RegisterTagNameFunc")

```
func (v *Validate) RegisterTagNameFunc(fn TagNameFunc)
```

RegisterTagNameFunc registers a function to get alternate names for StructFields.

eg. to use the names which have been specified for JSON representations of structs, rather than normal Go field names:

```
validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
    name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
    // skip if tag key says it should be ignored
    if name == "-" {
        return ""
    }
    return name
})
```

#### func (\*Validate) [RegisterTranslation](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L321) [¶](#Validate.RegisterTranslation "Go to Validate.RegisterTranslation")

```
func (v *Validate) RegisterTranslation(tag string, trans ut.Translator, registerFn RegisterTranslationsFunc, translationFn TranslationFunc) (err error)
```

RegisterTranslation registers translations against the provided tag.

#### func (\*Validate) [RegisterValidation](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L220) [¶](#Validate.RegisterValidation "Go to Validate.RegisterValidation")

```
func (v *Validate) RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error
```

RegisterValidation adds a validation with the given tag

NOTES:
- if the key already exists, the previous validation function will be replaced.
- this method is not thread-safe it is intended that these all be registered prior to any validation

#### func (\*Validate) [RegisterValidationCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L226) [¶](#Validate.RegisterValidationCtx "Go to Validate.RegisterValidationCtx")

```
func (v *Validate) RegisterValidationCtx(tag string, fn FuncCtx, callValidationEvenIfNull ...bool) error
```

RegisterValidationCtx does the same as RegisterValidation on accepts a FuncCtx validation
allowing context.Context validation support.

#### func (\*Validate) [SetTagName](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L158) [¶](#Validate.SetTagName "Go to Validate.SetTagName")

```
func (v *Validate) SetTagName(name string)
```

SetTagName allows for changing of the default tag name of 'validate'

#### func (\*Validate) [Struct](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L345) [¶](#Validate.Struct "Go to Validate.Struct")

```
func (v *Validate) Struct(s interface{}) error
```

Struct validates a structs exposed fields, and automatically validates nested structs, unless otherwise specified.

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L354) [¶](#Validate.StructCtx "Go to Validate.StructCtx")

```
func (v *Validate) StructCtx(ctx context.Context, s interface{}) (err error)
```

StructCtx validates a structs exposed fields, and automatically validates nested structs, unless otherwise specified
and also allows passing of context.Context for contextual validation information.

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructExcept](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L522) [¶](#Validate.StructExcept "Go to Validate.StructExcept")

```
func (v *Validate) StructExcept(s interface{}, fields ...string) error
```

StructExcept validates all fields except the ones passed in.
Fields may be provided in a namespaced fashion relative to the struct provided
i.e. NestedStruct.Field or NestedArrayField[0].Struct.Name

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructExceptCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L533) [¶](#Validate.StructExceptCtx "Go to Validate.StructExceptCtx")

```
func (v *Validate) StructExceptCtx(ctx context.Context, s interface{}, fields ...string) (err error)
```

StructExceptCtx validates all fields except the ones passed in and allows passing of contextual
validation information via context.Context
Fields may be provided in a namespaced fashion relative to the struct provided
i.e. NestedStruct.Field or NestedArrayField[0].Struct.Name

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructFiltered](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L389) [¶](#Validate.StructFiltered "Go to Validate.StructFiltered")

```
func (v *Validate) StructFiltered(s interface{}, fn FilterFunc) error
```

StructFiltered validates a structs exposed fields, that pass the FilterFunc check and automatically validates
nested structs, unless otherwise specified.

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructFilteredCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L399) [¶](#Validate.StructFilteredCtx "Go to Validate.StructFilteredCtx")

```
func (v *Validate) StructFilteredCtx(ctx context.Context, s interface{}, fn FilterFunc) (err error)
```

StructFilteredCtx validates a structs exposed fields, that pass the FilterFunc check and automatically validates
nested structs, unless otherwise specified and also allows passing of contextual validation information via
context.Context

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructPartial](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L436) [¶](#Validate.StructPartial "Go to Validate.StructPartial")

```
func (v *Validate) StructPartial(s interface{}, fields ...string) error
```

StructPartial validates the fields passed in only, ignoring all others.
Fields may be provided in a namespaced fashion relative to the struct provided
eg. NestedStruct.Field or NestedArrayField[0].Struct.Name

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [StructPartialCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L447) [¶](#Validate.StructPartialCtx "Go to Validate.StructPartialCtx")

```
func (v *Validate) StructPartialCtx(ctx context.Context, s interface{}, fields ...string) (err error)
```

StructPartialCtx validates the fields passed in only, ignoring all others and allows passing of contextual
validation information via context.Context
Fields may be provided in a namespaced fashion relative to the struct provided
eg. NestedStruct.Field or NestedArrayField[0].Struct.Name

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.

#### func (\*Validate) [ValidateMap](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L194) [¶](#Validate.ValidateMap "Go to Validate.ValidateMap") added in v10.6.0

```
func (v *Validate) ValidateMap(data map[string]interface{}, rules map[string]interface{}) map[string]interface{}
```

ValidateMap validates map data from a map of tags

#### func (Validate) [ValidateMapCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L164) [¶](#Validate.ValidateMapCtx "Go to Validate.ValidateMapCtx") added in v10.6.0

```
func (v Validate) ValidateMapCtx(ctx context.Context, data map[string]interface{}, rules map[string]interface{}) map[string]interface{}
```

ValidateMapCtx validates a map using a map of validation rules and allows passing of contextual
validation information via context.Context.

#### func (\*Validate) [Var](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L593) [¶](#Validate.Var "Go to Validate.Var")

```
func (v *Validate) Var(field interface{}, tag string) error
```

Var validates a single variable using tag style validation.
eg.
var i int
validate.Var(i, "gt=1,lt=10")

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### func (\*Validate) [VarCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L611) [¶](#Validate.VarCtx "Go to Validate.VarCtx")

```
func (v *Validate) VarCtx(ctx context.Context, field interface{}, tag string) (err error)
```

VarCtx validates a single variable using tag style validation and allows passing of contextual
validation information via context.Context.
eg.
var i int
validate.Var(i, "gt=1,lt=10")

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### func (\*Validate) [VarWithKey](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L697) [¶](#Validate.VarWithKey "Go to Validate.VarWithKey") added in v10.28.0

```
func (v *Validate) VarWithKey(key string, field interface{}, tag string) error
```

VarWithKey validates a single variable with a key to be included in the returned error using tag style validation
eg.
var s string
validate.VarWithKey("email\_address", s, "required,email")

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### func (\*Validate) [VarWithKeyCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L715) [¶](#Validate.VarWithKeyCtx "Go to Validate.VarWithKeyCtx") added in v10.28.0

```
func (v *Validate) VarWithKeyCtx(ctx context.Context, key string, field interface{}, tag string) (err error)
```

VarWithKeyCtx validates a single variable with a key to be included in the returned error using tag style validation
and allows passing of contextual validation information via context.Context.
eg.
var s string
validate.VarWithKeyCtx("email\_address", s, "required,email")

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### func (\*Validate) [VarWithValue](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L646) [¶](#Validate.VarWithValue "Go to Validate.VarWithValue")

```
func (v *Validate) VarWithValue(field interface{}, other interface{}, tag string) error
```

VarWithValue validates a single variable, against another variable/field's value using tag style validation
eg.
s1 := "abcd"
s2 := "abcd"
validate.VarWithValue(s1, s2, "eqcsfield") // returns true

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### func (\*Validate) [VarWithValueCtx](https://github.com/go-playground/validator/blob/v10.30.1/validator_instance.go#L665) [¶](#Validate.VarWithValueCtx "Go to Validate.VarWithValueCtx")

```
func (v *Validate) VarWithValueCtx(ctx context.Context, field interface{}, other interface{}, tag string) (err error)
```

VarWithValueCtx validates a single variable, against another variable/field's value using tag style validation and
allows passing of contextual validation information via context.Context.
eg.
s1 := "abcd"
s2 := "abcd"
validate.VarWithValue(s1, s2, "eqcsfield") // returns true

WARNING: a struct can be passed for validation eg. time.Time is a struct or
if you have a custom type and have registered a custom type handler, so must
allow it; however unforeseen validations will occur if trying to validate a
struct that is meant to be passed to 'validate.Struct'

It returns InvalidValidationError for bad values passed in and nil or ValidationErrors as error otherwise.
You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
validate Array, Slice and maps fields which may contain more than one error

#### type [ValidationErrors](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L36) [¶](#ValidationErrors "Go to ValidationErrors")

```
type ValidationErrors []FieldError
```

ValidationErrors is an array of FieldError's
for use in custom error messages post validation.

#### func (ValidationErrors) [Error](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L42) [¶](#ValidationErrors.Error "Go to ValidationErrors.Error")

```
func (ve ValidationErrors) Error() string
```

Error is intended for use in development + debugging and not intended to be a production error message.
It allows ValidationErrors to subscribe to the Error interface.
All information to create an error message specific to your application is contained within
the FieldError found within the ValidationErrors array

#### func (ValidationErrors) [Translate](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L54) [¶](#ValidationErrors.Translate "Go to ValidationErrors.Translate")

```
func (ve ValidationErrors) Translate(ut ut.Translator) ValidationErrorsTranslations
```

Translate translates all of the ValidationErrors

#### type [ValidationErrorsTranslations](https://github.com/go-playground/validator/blob/v10.30.1/errors.go#L17) [¶](#ValidationErrorsTranslations "Go to ValidationErrorsTranslations")

```
type ValidationErrorsTranslations map[string]string
```

ValidationErrorsTranslations is the translation return type
