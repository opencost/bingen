package java

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencost/bingen/pkg/util"
	"github.com/opencost/bingen/tests/streamfeed"
)

// goldenBuffer writes a fixed schema of every primitive + string + raw bytes
// using the reference go util.Buffer. The java harness in bufferParityHarness
// reads this exact schema in order, so the two must stay in lockstep.
func goldenBuffer() []byte {
	b := util.NewBuffer()
	b.WriteBool(true)
	b.WriteInt(-123456)
	b.WriteInt8(-8)
	b.WriteInt16(-1600)
	b.WriteInt32(-32000000)
	b.WriteInt64(-9000000000000)
	b.WriteUInt(4000000000)
	b.WriteUInt8(200)
	b.WriteUInt16(60000)
	b.WriteUInt32(4000000001)
	b.WriteUInt64(uint64(0xF0F1F2F3F4F5F6F7))
	b.WriteFloat32(3.5)
	b.WriteFloat64(-12345.6789)
	b.WriteString("hello")
	b.WriteString("héllo✓")
	b.WriteBytes([]byte{1, 2, 3, 255, 0})
	return b.Bytes()
}

// bufferParityHarness reads the go-produced golden blob with the generated
// ReadBuffer, spot-checks a handful of values, then re-encodes every value with
// the generated WriteBuffer and asserts the bytes match the golden blob exactly.
// It needs no expected values baked in beyond the spot-checks: read∘write must
// reproduce the reference encoder's bytes.
const bufferParityHarness = `import com.bingen.ReadBuffer;
import com.bingen.WriteBuffer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Arrays;

public final class BufferParity {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) throws Exception {
        byte[] golden = Files.readAllBytes(Path.of(args[0]));
        ReadBuffer rb = ReadBuffer.fromBytes(golden);

        boolean v1 = rb.readBool();
        int v2 = rb.readInt();
        byte v3 = rb.readInt8();
        short v4 = rb.readInt16();
        int v5 = rb.readInt32();
        long v6 = rb.readInt64();
        long v7 = rb.readUInt();
        int v8 = rb.readUInt8();
        int v9 = rb.readUInt16();
        long v10 = rb.readUInt32();
        long v11 = rb.readUInt64();
        float v12 = rb.readFloat32();
        double v13 = rb.readFloat64();
        String v14 = rb.readString();
        String v15 = rb.readString();
        byte[] v16 = rb.readBytes(5);

        check(v1, "readBool");
        check(v2 == -123456, "readInt");
        check(v3 == (byte) -8, "readInt8");
        check(v4 == (short) -1600, "readInt16");
        check(v5 == -32000000, "readInt32");
        check(v6 == -9000000000000L, "readInt64");
        check(v7 == 4000000000L, "readUInt");
        check(v8 == 200, "readUInt8");
        check(v9 == 60000, "readUInt16");
        check(v10 == 4000000001L, "readUInt32");
        check(v11 == 0xF0F1F2F3F4F5F6F7L, "readUInt64");
        check(v12 == 3.5f, "readFloat32");
        check(v13 == -12345.6789, "readFloat64");
        check(v14.equals("hello"), "readString ascii");
        check(v15.equals("héllo✓"), "readString utf8");
        check(Arrays.equals(v16, new byte[] {1, 2, 3, (byte) 255, 0}), "readBytes");

        WriteBuffer wb = WriteBuffer.create();
        wb.writeBool(v1);
        wb.writeInt(v2);
        wb.writeInt8(v3);
        wb.writeInt16(v4);
        wb.writeInt32(v5);
        wb.writeInt64(v6);
        wb.writeUInt(v7);
        wb.writeUInt8(v8);
        wb.writeUInt16(v9);
        wb.writeUInt32(v10);
        wb.writeUInt64(v11);
        wb.writeFloat32(v12);
        wb.writeFloat64(v13);
        wb.writeString(v14);
        wb.writeString(v15);
        wb.writeBytes(v16);

        byte[] out = wb.bytes();
        check(Arrays.equals(out, golden), "re-encoded bytes != golden");
        System.out.println("OK");
    }
}
`

func TestRuntimeBufferParity(t *testing.T) {
	const goPackage = "parity"
	const javaPackage = "com.example"
	const javaHarness = "BufferParity.java"

	checkJava(t)

	javaMain := &javaSource{
		FileName: javaHarness,
		Source:   bufferParityHarness,
	}

	dir := createAndEmitRuntime(t, goPackage, javaPackage)
	readBuffSrc := filepath.Join("com", "bingen", "ReadBuffer.java")
	writeBuffSrc := filepath.Join("com", "bingen", "WriteBuffer.java")

	// write golden buffer to runtime dir
	goldenPath := filepath.Join(dir, "golden.bin")
	if err := os.WriteFile(goldenPath, goldenBuffer(), 0o600); err != nil {
		t.Fatalf("writing golden: %v", err)
	}

	// write java source to runtime dir
	writeJavaSource(t, dir, javaMain)

	compileJava(t, dir, []string{
		readBuffSrc,
		writeBuffSrc,
		javaMain.FileName,
	})

	out, err := runJavaTest(t, dir, javaMain.HarnessName(), "golden.bin")
	if err != nil {
		t.Fatalf("java buffer parity harness failed: %v\n%s", err, out)
	}
	t.Logf("%s", string(out))
}

// goStreamSeq streams data with the generated go streamer and returns the same
// canonical representation the Java harness produces.
func goStreamSeq(data []byte) []string {
	var seq []string
	stream := streamfeed.NewFeedStream(bytes.NewReader(data))
	defer stream.Close()

	for fi, v := range stream.Stream() {
		switch {
		case v == nil:
			seq = append(seq, fi.Name+"\t\t<nil>")
		case v.Index != nil:
			seq = append(seq, fmt.Sprintf("%s\t%v\t%v", fi.Name, v.Index, v.Value))
		default:
			seq = append(seq, fmt.Sprintf("%s\t\t%v", fi.Name, v.Value))
		}
	}
	return seq
}

// TestStreaming_GoJavaParity feeds identical encoded bytes to the Go and Java
// streamers and asserts they yield the same (field, index, value) sequence,
// proving the Java streamer reads the wire format exactly like Go's.
func TestStreaming_GoJavaParity(t *testing.T) {
	// streamSeqHarness streams the go-encoded bytes with the generated Java push
	// streamer and writes one canonical "field\tindex\tvalue" line per yield.
	const streamSeqHarness = `import com.opencost.streamfeed.FeedStreamer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

public final class StreamSeqMain {
    public static void main(String[] args) throws Exception {
        byte[] data = Files.readAllBytes(Path.of(args[0]));
        List<String> seq = new ArrayList<>();
        FeedStreamer.stream(data, (fi, v) -> {
            String line;
            if (v == null) {
                line = fi.name() + "\t\t<nil>";
            } else if (v.index() != null) {
                line = fi.name() + "\t" + String.valueOf(v.index()) + "\t" + String.valueOf(v.value());
            } else {
                line = fi.name() + "\t\t" + String.valueOf(v.value());
            }
            seq.add(line);
            return true;
        });
        Files.writeString(Path.of(args[1]), String.join("\n", seq));
    }
}
`

	const goPackage = "streamfeed"
	const harnessName = "StreamSeqMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   streamSeqHarness,
	}

	dir := buildJavaTest(t, goPackage, javaMain)
	inPath := filepath.Join(dir, "feed.bin")
	seqPath := filepath.Join(dir, "javaseq.txt")

	feed := &streamfeed.Feed{
		Name:  "news",
		Tags:  []string{"a", "b", "c"},
		Count: 7,
		Meta:  map[string]string{"k1": "v1", "k2": "v2"},
	}
	data, err := feed.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	if err := os.WriteFile(inPath, data, 0o600); err != nil {
		t.Fatalf("writing feed.bin: %v", err)
	}

	o, err := runJavaTest(t, dir, harnessName, "feed.bin", "javaseq.txt")
	if err != nil {
		t.Fatalf("StreamSeqMain failed: %v\n%s", err, o)
	}

	javaSeq, err := os.ReadFile(seqPath)
	if err != nil {
		t.Fatalf("reading java sequence: %v", err)
	}

	goSeq := strings.Join(goStreamSeq(data), "\n")
	if goSeq != string(javaSeq) {
		t.Fatalf("stream sequence mismatch:\n go:\n%s\n java:\n%s", goSeq, javaSeq)
	}

	// Sanity: the sequence must actually contain the flattened elements.
	if !strings.Contains(goSeq, "Tags\t0\ta") || !strings.Contains(goSeq, "Meta\tk") {
		t.Fatalf("unexpected sequence content:\n%s", goSeq)
	}
}
