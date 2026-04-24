# bingen
Binary Codec Generator for annotated structs in go.

### Install
Using an ssh-agent and git, issue a global config update:
```bash
$ git config --global url.git@github.com:.insteadOf https://github.com/
```

Then get `bingen`:
```bash
$ go get github.com/opencost/bingen/cmd/bingen
```

Then install `bingen`:
```bash
$ go install github.com/opencost/bingen/cmd/bingen
```

### Usage
```
Usage of bingen:
        bingen [flags] -package P [directory]
Flags:
  -buffer string
        qualified package for the Buffer type (default "github.com/opencost/bingen/pkg/util")
  -package string
        package name to generate binary codecs for
  -version uint8
        the versioning to use for the default version set (default 1)
```

##### Buffer
The buffer flag should point to the location of the `util.Buffer` type. Since this is currently a private repository, it's best to just copy/paste https://github.com/opencost/bingen/blob/develop/pkg/util/buffer.go into a `pkg/util` within your project. For instance, let's say you copy `buffer.go` to `pkg/util` in your project `github.com/bruh/gen-test`, then the buffer flag would be passed as `-buffer=github.com/bruh/gen-test/pkg/util`

##### Example
The easiest way to use `bingen` is via `go:generate`. In a project that contains custom struct types you wish to generate `MarshalBinary` and `UnmarshalBinary` methods for, navigate to the target package. Assuming that the package `pkg/stuff` has 3 types you want to generate binary marshal/unmarshal for: `Foo`, `Bar`, and `Widget`, create a new source file in `pkg/stuff` with the following:
```go
package stuff

// @bingen:generate:Foo
// @bingen:generate:Bar
// @bingen:generate:Widget

//go:generate bingen -package=stuff -version=1 -buffer=github.com/bruh/gen-test/pkg/util
```

If you're using VSCode, a link will appear above `//go:generate ...`. Click the `run go generate ./...` option. This should run and create a `stuff_codecs.go` source file in the `pkg/stuff` directory. 

##### Non Standard Library Types
If you're using a non-standard library type as a field on a type targetted for generation, you'll need to ensure that the type implements both `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`. In addition, the type's import path will need to be output in the generated output. To accomplish this, you can use the `// @bingen:import:<package import>` directive. For example, if you're using the package: "github.com/acme/widgets/pkg/widget", you would need to add the following to anywhere in the target package:

```go
// @bingen:import:github.com/acme/widgets/pkg/widget
```

##### External Alias Types
A more advanced version of the *import* command is the *define* command. This is used when you have an alias type that you want to be treated as a first class citizen in the generated code. For example, let's say you have `type WidgetID string` in a `shared` package, and you want to used `shared.WidgetID` on a field within your bingen package. If you were to only use the `// @bingen:import:github.com/acme/widgets/pkg/shared` directive, then the generated code would treat `shared.WidgetID` as a `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` implementation, not as a simple string. In order to have the generated code treat `shared.WidgetID` as a string, you need to define the external alias type with the `define` directive:

Let's say we have the following `WidgetID` alias type in the `github.com/acme/widgets/pkg/shared` package:
```go 
package shared

type WidgetID string 
```

Now, in our target bingen package, we have the following type that uses `shared.WidgetID`:
```go 
package stuff 

import "github.com/acme/widgets/pkg/shared"

type Widget struct {
      ID shared.WidgetID 
}
```

Now our bingen syntax will need to include the `define` directive for `shared.WidgetID`:
```go
package stuff

// @bingen:define[string]:github.com/acme/widgets/pkg/shared.WidgetID

// @bingen:generate:Widget

//go:generate bingen -package=stuff -version=1 -buffer=github.com/acme/widgets/pkg/util 
```

Note that the `define` directive also implicitly imports the package, so you do not need to also include an `import` directive for the same package. To summarize, the `import` directive is used when you have a non-standard library type that implements `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`, and you want to use it as a field on a generated type. The `define` directive is used when you have an alias type that does not implement `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`, but you want to treat it as a first class citizen in the generated code with the underlying type's encoding/decoding behavior.


##### Interface Implementations
It's important that if you type any fields as an `interface`, you'll need to annotate both the interface type as well as any implementations of that type for generation. Continuing the example above, assume we have a `type Thing interface` which both `Bar`, `Widget`, and a new type `Gear` implement. We'll need to add the annotation for both the new type and the interface:

```go
package stuff

// @bingen:generate:Foo
// @bingen:generate:Bar
// @bingen:generate:Widget
// @bingen:generate:Gear
// @bingen:generate:Thing


//go:generate bingen -package=stuff -version=2 -buffer=github.com/bruh/gen-test/pkg/util
```

##### String Table 
There exists a string table implementation which allows nested objects to leverage a shared set of strings in the encodings. For larger resources that contain many repetitive strings, this will be the most optimal option. However, the impact will be less noticeable on smaller resources. To apply a string table encoding/decoding, use the `generate` command options via `[]`:

```go
package stuff

// @bingen:generate:Foo
// @bingen:generate[stringtable]:Bar
// @bingen:generate:Widget
// @bingen:generate:Gear
// @bingen:generate:Thing


//go:generate bingen -package=stuff -version=3 -buffer=github.com/bruh/gen-test/pkg/util
```

This option must be applied to a concrete type. 

##### Version Sets
In the event that you have a resources that should be versioned separately, you can apply a `VersionSet` to generation. This is an ordered operation that needs to be annotated prior to `generate` annotations you wish to be added to that set. The syntax for a version set:

```go
// @bingen:set[name=<version-set-name>,version=<set-version>]
// @bingen:generate:ResourceInVersionSet
// ...
// @bingen:end
```

Note that the `end` command pops the scope of the version set. The following applies version sets for the previous example:

```go
package stuff

// Any generates that appear outside of a version set scope will be applied in the default version set,
// which uses the value passed via -version flag (3 in this example)
// @bingen:generate:User

// @bingen:set[name=FooBar,version=3]
// @bingen:generate:Foo
// @bingen:generate[stringtable]:Bar
// @bingen:end

// @bingen:set[name=Widgets,version=4]
// @bingen:generate:Widget
// @bingen:generate:Gear
// @bingen:generate:Thing
// @bingen:end


//go:generate bingen -package=stuff -version=3 -buffer=github.com/bruh/gen-test/pkg/util
```

##### Ignoring Fields
You can also ignore fields in the generation process. This is useful if you have a transient field you wish to be ignored during marshal and unmarshal. This is done by adding the `@bingen:field[ignore]` annotation directly after the field:

```go
type Person struct {
      FirstName string 
      LastName string
      FullName string //@bingen:field[ignore]
}
```

In this example, `FirstName` and `LastName` will both be marshalled and unmarshalled. `FullName` will be ignored.

##### Pre and Post Processing Types
You can apply pre and post processing hooks to any generated type. These hooks often work well with ignored/transient fields to ensure that any data being encoded can be populated from the transient data. Likewise, you can also ensure that any transient fields are populated on decode as well. These hooks are enabled via the `generate` options:

```go
// @bingen:generate[stringtable,preprocess,postprocess]:Person
```

Using the `Person` from the previous example, let's have the pre and post processor functions manage the `FullName` field. It's important to note that when you mark a type with `preprocess` and/or `postprocess`, you must also ensure the existence of two functions in the package you have targetting for generation: 

For the `preprocess`, the function name will be `preProcess<Type>(myType *<Type>)`. For our `Person` example, it would look like this:
```go
func preProcessPerson(p *Person) {
      // If the FullName field is set, update the FirstName and LastName fields
      if p.FullName != "" {
            firstAndLast := strings.Split(p.FullName, " ")
            // Make FullName contains a first and last name separated by space
            if len(firstAndLast) != 2 {
                  return
            }

            p.FirstName = firstAndLast[0]
            p.LastName = firstAndLast[1]
      }
}
```

For the `postprocess`, the function name will be `postProcess<Type>(myType *<Type>)`. For our `Person` example, it would look like this:
```go
func postProcessPerson(p *Person) {
      // Set the FullName field to the concatenation of FirstName and LastName separated by a space
      p.FullName = fmt.Sprintf("%s %s", p.FirstName, p.LastName)
}
```

##### Migration of Types
Similar to the pre and post processing hooks for generated types, you can also specify a migration hook. A migration hook is used when a higher versioned struct unmarshals from a lesser versioned encoding. The most common used of this feature would be to load older data, make a one time change, then store out the new result data. This hook is enabled via the `generate` options:

```go
// @bingen:generate[stringtable,preprocess,postprocess,migrate]:Person
```

For the `migrate` option, the function name will be `migrate<Type>(myType *<Type>, fromVersion uint8, toVersion uint8)`. For our `Person` example, it would look like this:
```go
func migratePerson(p *Person, fromVersion uint8, toVersion uint8) {
      if fromVersion == 1 && toVersion == 3 {
            // special handling for a new field added in v3, loaded from v1 
      }
      if fromVersion == 2 && toVersion == 3 {
            // special handling for a new field added in v3, loaded from v2 
      }
}
```


##### Streamable Types
The `streamable` option generates an `io.Reader`-based streaming iterator for a type, allowing its fields to be decoded one at a time without unmarshalling the entire object into memory. Slices and maps are flattened one level deep, with each element yielded individually. This is well suited for large types where only a subset of fields need to be inspected, or where memory pressure makes loading the full object at once undesirable.

To enable streaming for a type, add `streamable` to the `generate` annotation options:

```go
package stuff

// @bingen:generate[streamable]:Foo

//go:generate bingen -package=stuff -version=1 -buffer=github.com/bruh/gen-test/pkg/util
```

For each `streamable` type, bingen generates a `<Type>Stream` struct implementing the `BingenStream` interface. The interface exposes three methods: `Stream()`, which returns an `iter.Seq2[BingenFieldInfo, *BingenValue]` iterator; `Close()`, which releases the underlying `io.Reader`; and `Error()`, which returns any error that occurred during streaming and should be checked after iteration completes.

`NewStreamFor[T]` is a generated generic function that creates the appropriate stream from a given `io.Reader` for type `T`. It returns an error if `T` was not annotated with `streamable`. Each iteration of the stream yields a `BingenFieldInfo` with the field's `Name string` and `Type reflect.Type`, along with a `*BingenValue` carrying the decoded `Value`. A `nil` `*BingenValue` (detected via `IsNil()`) indicates that a nilable field was encoded as `nil`. For slice and map fields, `BingenValue.Index` holds the element's integer index or map key respectively.

Using the `Foo` type from the previous example, the following reads its fields from an `io.Reader`:

```go
stream, err := stuff.NewStreamFor[stuff.Foo](reader)
if err != nil {
    fmt.Printf("Error creating stream: %s\n", err)
    return
}
defer stream.Close()

for fi, bv := range stream.Stream() {
    if bv.IsNil() {
        fmt.Printf("Field: %s (nil)\n", fi.Name)
        continue
    }
    // For slice or map fields, bv.Index holds the element's index or key.
    if bv.Index != nil {
        fmt.Printf("Field: %s[%v] = %v\n", fi.Name, bv.Index, bv.Value)
        continue
    }
    fmt.Printf("Field: %s = %v\n", fi.Name, bv.Value)
}

if err := stream.Error(); err != nil {
    fmt.Printf("Stream error: %s\n", err)
}
```

The `streamable` option can be combined with other options such as `stringtable`:

```go
// @bingen:generate[streamable,stringtable]:Foo
```

When streaming a type that was encoded with a string table, the stream automatically reads and resolves the string table from the `io.Reader` before yielding any fields. For memory-constrained environments, bingen can spill the string table to a temporary file rather than holding it entirely in memory. This is configured once per package via `ConfigureBingen`:

```go
stuff.ConfigureBingen(&stuff.BingenConfiguration{
    FileBackedStringTableEnabled: true,
    FileBackedStringTableDir:     "/tmp",
})
```

When `FileBackedStringTableEnabled` is `true`, the string table is written to a temporary file in `FileBackedStringTableDir` and individual strings are read from disk on demand. This reduces peak memory usage at the cost of additional file I/O, and pairs well with streaming reads for high-throughput processing of large binary datasets.

#### Backwards Compatibility and Field Versioning
Bingen supports backwards compatibility, but it depends on the user to annotate new fields with the first version the field appears and any specific default values (optional). 

For example, let's say we have a struct `Container`:
```go
type Container struct {
      Name     string 
      Children []string 
}
```

and in our `bingen.go` file, we setup a version set:
```go
// @bingen:set[name=ContainerExample,version=1]
// @bingen:generate:Container
// @bingen:end

//go:generate bingen -package=container -version=1 -buffer=github.com/container-example/pkg/util
```

Now, if we generate our codec, we can write code that marshals a `Container` instance:
```go
c := &Container{
      Name: "TestContainer",
      Children: []string{
            "Child1",
            "Child2",
            "Child3",
      },
}

b, err := c.MarshalBinary()
if err != nil {
      fmt.Printf("Error: %s\n", err)
      return 
}

// Write Encoded Binary out to a File
err = os.WriteFile("container.bin", b, 0644)
if err != nil {
      fmt.Printf("Failed to save container.bin: %s\n", err)
      return 
}
```

Now, some time later, we want to update our `Container` struct with a new property: `Parent string`. If we simply add the field and update the version, then our old saved binary file `container.bin` will fail to unmarshal as the new `Container` type. 

However, we can ensure that bingen knows that the new field is specific to the next version by annotating the specific _new_ fields on the `Container`. These annotations are different than other `//@bingen` tags in that they are specific to the field in which they're applied. 

To annotate our new `Parent string` field, we'll add it to `Container` and use the field versioning annotation:
```go
type Container struct {
      Name     string 
      Children []string 
      Parent   string    // @bingen:field[version=2]
}
```

We then need to ensure our version set version is also updated:
```go
// @bingen:set[name=ContainerExample,version=2]
// @bingen:generate:Container
// @bingen:end

//go:generate bingen -package=container -version=2 -buffer=github.com/container-example/pkg/util
```

Now if we were to load our file `container.bin`:
```go
file, err := os.ReadFile("container.bin")
if err != nil {
      fmt.Printf("Error reading file container.bin: %s\n", err)
      return
}

c := &Container{}
err = c.UnmarshalBinary(file)
if err != nil {
      fmt.Printf("Failed to unmarshal binary: %s\n", err)
      return
}

fmt.Printf("Name: %s\n", c.Name)
for _, child := range c.Children {
      fmt.Printf("Child: %s\n", child)
}
// Now we check the new property, Parent:
fmt.Printf("Parent: %s\n", c.Parent)

// outputs: 
// Name: TestContainer
// Child: Child1
// Child: Child2
// Child: Child3
// Parent: 
```

Note that the `Parent` was set to the empty string when we unmarshalled the old binary format. This is the default for a `string`, but if you need a specific default to be set when loading older versions, you may specify in the `//@bingen:field` annotation:

```go
type Container struct {
      Name     string 
      Children []string 
      Parent   string    // @bingen:field[version=2, default=momma-container]
}
```

Now if we were to re-run the `ReadFile` example, the output would be:
```go
// outputs: 
// Name: TestContainer
// Child: Child1
// Child: Child2
// Child: Child3
// Parent: momma-container
```

##### Defaults 
While the `default` is optional in the field version annotation, there are a few special cases to be aware of:
* Any nilable type will default to `nil` and ignore the `default` value set. This include `map`, `slice`, `interface` types, and pointer types. 
* String fields default value does not need to include `"` characters (see above example)
* Alias types may require casting and more explicit defaults. For example, if you alias `type Foo int`, then use a property `F Foo`, your tag may need to cast the default: 
```go
F Foo //@bingen:field[version=2, default=Foo(15)]
```

Now, let's assume we want to add another field to our `Container` type:
```go
type Container struct {
      Name     string 
      Children []string 
      Parent   string    // @bingen:field[version=2, default=none]
      Size     int       // @bingen:field[version=3, default=1]
}
```

We advance the version set as well:

```go
// @bingen:set[name=ContainerExample,version=3]
// @bingen:generate:Container
// @bingen:end

//go:generate bingen -package=container -version=3 -buffer=github.com/container-example/pkg/util
```

This does _NOT_ prevent us from loading the `container.bin` file that was generated using the v1 shema. Version 3 can unmarshal Version 2 and Version 1. 

