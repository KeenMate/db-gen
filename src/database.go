package dbGen

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"slices"
)

type DbConn struct {
	conn *sqlx.DB
}

func Connect(connectionString string) (*DbConn, error) {
	connection, err := sqlx.Connect("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	err = connection.Ping()

	if err != nil {
		return nil, err
	}

	conn := &DbConn{conn: connection}

	return conn, nil
}

func (conn DbConn) Select(output interface{}, query string, params ...interface{}) error {
	err := conn.conn.Select(output, query, params...)
	return err
}

type DbRoutine struct {
	RowNumber     int    `db:"row_number"`
	RoutineSchema string `db:"routine_schema"`
	RoutineName   string `db:"routine_name"`
	SpecificName  string `db:"specific_name"`
	DataType      string `db:"data_type"`
	UdtTypeScheme string `db:"type_udt_schema"`
	UdtTypeName   string `db:"type_udt_name"`
	ParamCount    int    `db:"param_count"`
	FuncType      string `db:"func_type"`
	InParameters  []DbParameter
	OutParameters []DbParameter
}

type DbParameter struct {
	OrdinalPosition int    `db:"ordinal_position"`
	Name            string `db:"parameter_name"`
	Mode            string `db:"parameter_mode"` // IN/OUT
	UDTName         string `db:"udt_name"`       // User defined type
	IsNullable      bool   `db:"is_nullable"`
}

const (
	OutMode   = "OUT"
	InMode    = "IN"
	Procedure = "procedure"
)

func GetRoutines(conn *DbConn, config *Config) ([]DbRoutine, error) {
	schemas := getSchemas(config)

	routines := make([]DbRoutine, 0)
	for _, schema := range schemas {
		newRoutines, err := getFunctionsInSchema(conn, schema)

		if err != nil {
			return nil, fmt.Errorf("getting routines for schema %s : %s", schema, err)
		}

		routines = append(routines, newRoutines...)
	}

	for i, routine := range routines {
		err := addParamsToRoutine(conn, &routines[i])

		if err != nil {
			return nil, fmt.Errorf("getting params for routine %s: %s", routine.RoutineName, err)
		}
	}

	return routines, nil

}
func getSchemas(config *Config) []string {
	schemas := make([]string, 0)
	for _, schemaConfig := range config.Generate {
		if !slices.Contains(schemas, schemaConfig.Schema) {
			schemas = append(schemas, schemaConfig.Schema)
		}
	}

	return schemas
}

func getFunctionsInSchema(conn *DbConn, schema string) ([]DbRoutine, error) {
	routines := new([]DbRoutine)

	// I am coalescing
	q := `select row_number() over (PARTITION BY routine_schema, routine_name),
	        r.routine_schema::text,
	        r.routine_name::text,
	        r.specific_name::text,
	        coalesce(r.data_type,'') as data_type,
	        coalesce(r.type_udt_schema::text,'') as type_udt_schema,
	        coalesce(r.type_udt_name::text,'') as type_udt_name,
	        coalesce(param_count,0) as param_count,
	        case when r.data_type is null  then 'procedure' else 'function' end as func_type
	        
	      from information_schema.routines r
	      left join (select specific_schema, specific_name, count(*) as param_count from
	        information_schema.parameters p
	        group by  specific_schema, specific_name) p on p.specific_schema = r.specific_schema and p.specific_name = r.specific_name
	      where r.specific_schema = $1
	      order by routine_schema, routine_name;
	`

	err := conn.Select(routines, q, schema)
	if err != nil {
		return nil, err
	}

	return *routines, nil
}

func addParamsToRoutine(conn *DbConn, routine *DbRoutine) error {
	q := `
		select ordinal_position::int,
			   parameter_name::text,
			   parameter_mode::text,
			   udt_name::text,
			   parameter_default is not null as is_nullable
		from information_schema.parameters
		where specific_schema = $1
		  and specific_name = $2		
		union
		select c.ordinal_position::int, c.column_name::text, 'OUT', c.udt_name::text, c.column_default is not null
		from information_schema.columns c
		where c.table_name = $3
		  and c.table_schema = coalesce($4, 'public')
		union
		select a.ordinal_position::int, a.attribute_name::text, 'OUT', a.attribute_udt_name::text, is_nullable = 'YES'
		from information_schema.attributes a
		where a.udt_name = $3
		  and a.udt_schema = coalesce($4, 'public')
		order by ordinal_position;`

	params := new([]DbParameter)

	err := conn.Select(params, q, routine.RoutineSchema, routine.SpecificName, routine.UdtTypeName, routine.UdtTypeScheme)

	if err != nil {
		return err
	}

	for _, param := range *params {

		switch param.Mode {
		case InMode:
			routine.InParameters = append(routine.InParameters, param)
			break
		case OutMode:
			routine.OutParameters = append(routine.OutParameters, param)
			break
		}
	}
	return nil
}
