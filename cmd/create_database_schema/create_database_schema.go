package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/randomizedcoder/gdp/pkg/gdp"
	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"google.golang.org/protobuf/proto"
)

const (
	debugLevelCst        = 11
	signalChannelSizeCst = 10

	//promProtoFileCst = "/prometheus.proto"
	promProtoFileCst     = "/prometheus_protolist.proto"
	promListProtoFileCst = "/prometheus_protolist.proto"

	outFolderCst = "./out/"
)

var (
	debugLevel uint
)

func main() {

	of := flag.String("of", outFolderCst, "output folder")
	d := flag.Uint("d", debugLevelCst, "debug level")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	debugLevel = *d

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	complete := make(chan struct{}, signalChannelSizeCst)
	go initSignalHandler(cancel, complete)

	g := &gdp.GDP{}

	g.Config = &gdp_config.GdpConfig{
		PromProtoFile:     promProtoFileCst,
		PromListProtoFile: promListProtoFileCst,
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	g.InitMarshalConfigs(wg)

	wg.Wait()

	protoType := &gdpp.Envelope_PromRecordCounter{}

	columnsSQL := generateTableColumnsSQL(protoType)

	if debugLevel > 10 {
		log.Println("---------------------------")
		log.Println(columnsSQL)
		log.Println("---------------------------")
	}

	var select_count_from_tables strings.Builder

	g.MarshalConfigs.Range(func(key, value interface{}) bool {

		mc := value.(*gdp_config.MarshalConfig)

		if debugLevel > 10 {
			log.Println("****************************************")
			log.Printf("mc:%+v", mc)
		}

		tableName := fmt.Sprintf("gdp.%s", mc.Topic)

		mergeTreeSQL := generateCreateTableSQL(tableName, columnsSQL)
		kafkaSQL := generateCreateKafkaTableSQL(mc, tableName, columnsSQL)
		materializedViewSQL := generateCreateMaterializedViewSQL(tableName)

		if debugLevel > 10 {
			log.Println(mergeTreeSQL)
			log.Println("========================")
			log.Println(kafkaSQL)
			log.Println("========================")
			log.Println(materializedViewSQL)
		}

		baseFilename := fmt.Sprintf(*of+"%s", strings.ReplaceAll(mc.Topic, ".", "_"))

		err := os.WriteFile(baseFilename+".sql", []byte(mergeTreeSQL), 0644)
		if err != nil {
			log.Printf("Failed to write MergeTree SQL to file %s.sql: %v", baseFilename, err)
		} else {
			log.Printf("Wrote MergeTree SQL to file: %s.sql", baseFilename)
		}

		err = os.WriteFile(baseFilename+"_kafka.sql", []byte(kafkaSQL), 0644)
		if err != nil {
			log.Printf("Failed to write Kafka SQL to file %s_kafka.sql: %v", baseFilename, err)
		} else {
			log.Printf("Wrote Kafka SQL to file: %s_kafka.sql", baseFilename)
		}

		err = os.WriteFile(baseFilename+"_mv.sql", []byte(materializedViewSQL), 0644)
		if err != nil {
			log.Printf("Failed to write Materialized View SQL to file %s_mv.sql: %v", baseFilename, err)
		} else {
			log.Printf("Wrote Materialized View SQL to file: %s_mv.sql", baseFilename)
		}

		select_count_from_tables.WriteString(fmt.Sprintf("SELECT count(*) FROM %s;\n", tableName))

		return true
	})

	err := os.WriteFile(*of+"select_count_from_tables.sql", []byte(select_count_from_tables.String()), 0644)
	if err != nil {
		log.Printf("Failed to write select_count_from_tables.sql: %v", err)
	} else {
		log.Printf("Wrote select_count_from_tables.sql")
	}

}

// initSignalHandler sets up signal handling for the process, and
// will call cancel() when received
func initSignalHandler(cancel context.CancelFunc, complete <-chan struct{}) {

	c := make(chan os.Signal, signalChannelSizeCst)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Printf("Signal caught, closing application")
	cancel()

	log.Printf("Signal caught, cancel() called, and sleeping to allow goroutines to close")

	select {
	case <-complete:
		log.Printf("<-complete exit(0)")
	default:
		log.Printf("Sleep complete, goodbye! exit(0)")
	}

	os.Exit(0)
}

func generateTableColumnsSQL(protoType proto.Message) string {

	var fieldTypeOverrides = map[string]string{
		"TimestampNs": "DateTime64(9,'UTC') CODEC(DoubleDelta, LZ4)",
	}

	// tableName := fmt.Sprintf("gdp.%s", strings.ToLower(mc.Topic))

	var sb strings.Builder

	// sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))

	// Use reflection to iterate over the fields of the protobuf message.
	v := reflect.ValueOf(protoType).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		// Skip fields that are not exported or are internal to protobuf.
		if !field.IsExported() || strings.HasPrefix(fieldName, "XXX_") {
			continue
		}

		var clickhouseType string
		var ok bool

		// Check if there's an override for this field name.
		if overrideType, hasOverride := fieldTypeOverrides[fieldName]; hasOverride {
			clickhouseType = overrideType
			ok = true
		} else {
			// Otherwise, use the default type mapping.
			clickhouseType, ok = goTypeToClickHouseType(fieldType)
		}
		if !ok {
			log.Printf("Warning: Unsupported type %s for field %s, skipping", fieldType.String(), fieldName)
			continue
		}

		sb.WriteString(fmt.Sprintf("  %s %s,\n", fieldName, clickhouseType))
	}

	// Remove the trailing comma and newline.
	sql := sb.String()
	if debugLevel < 10 {
		log.Println(sql)
	}
	//sql = sql[:len(sql)-2] + "\n"
	// sql += ") ENGINE = MergeTree()\n"
	// sql += "ORDER BY (Timestamp_ns, Hostname, PollCounter, RecordCounter)\n"

	return sql
}

func generateCreateTableSQL(tableName string, columnsSQL string) string {

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("--\n"))
	sb.WriteString(fmt.Sprintf("-- %s.sql\n", tableName))
	sb.WriteString(fmt.Sprintf("--\n"))

	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n\n", tableName))

	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
	sb.WriteString(columnsSQL)
	sb.WriteString(")\n")
	sb.WriteString("ENGINE = MergeTree()\n")
	sb.WriteString("PARTITION BY toYYYYMM(TimestampNs)\n")
	sb.WriteString("ORDER BY (TimestampNs, Hostname, Pop, Label, Tag, PollCounter, RecordCounter)\n")
	sb.WriteString("TTL toDateTime(TimestampNs) + INTERVAL 14 DAY;\n\n")

	sb.WriteString(fmt.Sprintf("-- Note that ORDER BY clause implicitly specifies a primary key\n\n"))

	sb.WriteString(fmt.Sprintf("-- SHOW CREATE TABLE %s;\n", tableName))
	sb.WriteString(fmt.Sprintf("-- SELECT * FROM %s LIMIT 20;\n\n", tableName))

	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/guides/developer/ttl\n"))
	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/sql-reference/statements/alter/ttl\n"))
	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree#table_engine-mergetree-ttl\n"))
	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/sql-reference/functions/type-conversion-functions#todatetime\n\n"))

	sb.WriteString(fmt.Sprintf("-- end\n\n"))

	return sb.String()
}

func generateCreateKafkaTableSQL(mc *gdp_config.MarshalConfig, tableName string, columnsSQL string) string {

	filename := strings.TrimPrefix(mc.Filename, "/")

	var kafka_schema string
	switch mc.MarshalType {
	case "ProtobufSingle", "Protobuf":
		//kafka_schema = filename + ":" + "Envelope_PromRecordCounter"
		kafka_schema = filename + ":" + "PromRecordCounter"
	case "ProtobufList":
		//kafka_schema = filename + ":" + "Envelope"
		kafka_schema = filename + ":" + "PromRecordCounter"
		//kafka_schema = filename + ":" + "EnvelopePromRecordCounter"
	default:
		log.Fatal("unknown mc.MarshalType")
	}

	// https://clickhouse.com/docs/interfaces/formats/ProtobufList#example-usage

	var sb strings.Builder
	kafkaTableName := tableName + "_kafka"

	sb.WriteString(fmt.Sprintf("--\n"))
	sb.WriteString(fmt.Sprintf("-- %s.sql\n", kafkaTableName))
	sb.WriteString(fmt.Sprintf("--\n"))

	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n\n", kafkaTableName))

	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", kafkaTableName))
	sb.WriteString(columnsSQL)
	sb.WriteString(")\n")
	sb.WriteString("ENGINE = Kafka SETTINGS\n")
	sb.WriteString(fmt.Sprintf("  kafka_broker_list = 'redpanda-0:9092',\n"))
	sb.WriteString(fmt.Sprintf("  kafka_topic_list = '%s',\n", mc.Topic))
	sb.WriteString(fmt.Sprintf("  kafka_schema = '%s',\n", kafka_schema))
	sb.WriteString(fmt.Sprintf("  kafka_max_rows_per_message = 10000,\n"))

	sb.WriteString(fmt.Sprintf("  kafka_num_consumers = 1,\n"))
	sb.WriteString(fmt.Sprintf("  kafka_thread_per_consumer = 0,\n"))
	sb.WriteString(fmt.Sprintf("  kafka_group_name = '%s',\n", mc.Topic))
	sb.WriteString(fmt.Sprintf("  kafka_skip_broken_messages = 1,\n"))
	sb.WriteString(fmt.Sprintf("  kafka_handle_error_mode = 'stream',\n"))
	sb.WriteString(fmt.Sprintf("  kafka_format = '%s';\n\n", mc.MarshalType))
	// format must be last!
	// https://github.com/ClickHouse/ClickHouse/issues/37895

	sb.WriteString(fmt.Sprintf("-- SHOW CREATE TABLE %s;\n", kafkaTableName))
	sb.WriteString(fmt.Sprintf("-- SELECT * FROM system.kafka_consumers FORMAT Vertical;\n"))
	sb.WriteString(fmt.Sprintf("-- DETACH TABLE %s;\n", kafkaTableName))
	sb.WriteString(fmt.Sprintf("-- SELECT * FROM %s LIMIT 20;\n\n", kafkaTableName))

	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/integrations/kafka/kafka-table-engine\n"))
	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/engines/table-engines/integrations/kafka#creating-a-table\n"))
	sb.WriteString(fmt.Sprintf("-- kafka_format last! = https://github.com/ClickHouse/ClickHouse/issues/37895\n\n"))

	sb.WriteString(fmt.Sprintf("-- end\n\n"))

	return sb.String()
}

func generateCreateMaterializedViewSQL(tableName string) string {
	var sb strings.Builder
	materializedViewName := tableName + "_mv"
	kafkaTableName := tableName + "_kafka"

	sb.WriteString(fmt.Sprintf("--\n"))
	sb.WriteString(fmt.Sprintf("-- %s.sql\n", materializedViewName))
	sb.WriteString(fmt.Sprintf("--\n"))

	sb.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s;\n\n", materializedViewName))

	sb.WriteString(fmt.Sprintf("CREATE MATERIALIZED VIEW %s TO %s\n", materializedViewName, tableName))
	sb.WriteString(fmt.Sprintf("  AS SELECT *\n"))
	sb.WriteString(fmt.Sprintf("  FROM %s\n", kafkaTableName))
	sb.WriteString(fmt.Sprintf("  WHERE length(_error) == 0;\n\n"))

	sb.WriteString(fmt.Sprintf("-- SHOW CREATE TABLE %s;\n\n", materializedViewName))

	sb.WriteString(fmt.Sprintf("-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view\n\n"))

	sb.WriteString(fmt.Sprintf("-- See also:\n"))
	sb.WriteString(fmt.Sprintf("-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679\n"))
	sb.WriteString(fmt.Sprintf("-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py\n\n"))

	sb.WriteString(fmt.Sprintf("-- end\n\n"))

	return sb.String()
}

func goTypeToClickHouseType(t reflect.Type) (string, bool) {
	switch t.Kind() {
	case reflect.String:
		return "LowCardinality(String) CODEC(LZ4)", true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "UInt64 CODEC(DoubleDelta, LZ4)", true
	case reflect.Float32, reflect.Float64:
		//return "Float64 CODEC(LZ4)", true
		return "Float64", true
	case reflect.Bool:
		return "UInt8 CODEC(LZ4)", true
	case reflect.Slice:
		// Handle slices (e.g., []byte) as String for now.
		if t.Elem().Kind() == reflect.Uint8 {
			return "String CODEC(LZ4)", true
		}
		return "", false
	default:
		return "", false
	}
}
