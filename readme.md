# db-gen

### Configuration overview

- **ConnectionString (string)**:
    - Defines the PostgreSQL database connection string.
    - For example `postgresql://usernmae:password@localhost:5432/database_name`
- **OutputFolder (string)**:
    - Specifies the folder where generated code files will be saved.
    - It can be relative to current working directory
- **OutputNamespace (string)**:
    - Sets the namespace for the generated code.
- **GenerateModels (boolean)**:
    - If **True** Generates models
- **GenerateProcessors (boolean)**:
    - If **True** Generates processors
- **SkipModelGenForVoidReturns (boolean)**:
    - If **True** doesn't generate models for functions without return type
- **ClearOutputFolder (boolean)**:
    - If **True** deletes content of output folder before generating new files
- **DbContextTemplate (string)**:
    - Path to the template file for generating the dbContext file.
- **ModelTemplate (string)**:
    - Path to the template file for generating model file.
- **ProcessorTemplate (string)**:
    - Path to the template file for generating processor file.
- **GeneratedFileExtension (string)**:
    - Defines the file extension for generated files.
- **Generate**:
    - **Schema (string)**:
        - Specifies the database schema name.
    - **AllFunctions (boolean)**:
        - If true generated all functions except
    - **IgnoredFunctions (array of strings)**:
        - Functions to be ignored when generating code in the schema.
    - **Functions (array of strings)**:
        - Functions to be explicitly included when generating code in the schema.