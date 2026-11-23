package java

import (
	"testing"
)

// TestStreaming_Incremental proves the Java streamer consumes an InputStream
// incrementally (like Go's io.Reader), rather than requiring the whole payload
// in memory up front.
func TestStreaming_Incremental(t *testing.T) {
	const incrementalHarness = `import com.opencost.streamfeed.Feed;
import com.opencost.streamfeed.FeedEncoder;
import com.opencost.streamfeed.FeedStreamer;
import java.io.ByteArrayInputStream;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;

public final class IncrementalMain {
	static void check(boolean ok, String what) {
		if (!ok) {
			System.err.println("FAIL: " + what);
			System.exit(1);
		}
	}

	public static void main(String[] args) {
		List<String> tags = new ArrayList<>();
		for (int i = 0; i < 1000; i++) {
			tags.add("tag-" + i);
		}
		byte[] data = FeedEncoder.toBytes(new Feed("N", tags, 1, Map.of()));

		ByteArrayInputStream in = new ByteArrayInputStream(data);
		long[] consumedAtFirst = {-1};
		List<String> fields = new ArrayList<>();

		FeedStreamer.stream(in, (fi, v) -> {
			if (consumedAtFirst[0] < 0) {
				consumedAtFirst[0] = data.length - in.available();
			}
			fields.add(fi.name());
			return true;
		});

		// The first field (Name) must be delivered after consuming only a small
		// prefix of the stream -- proof it reads incrementally rather than
		// loading the whole payload first.
		check(consumedAtFirst[0] > 0 && consumedAtFirst[0] < data.length / 4,
			"first field delivered after consuming " + consumedAtFirst[0] + " of " + data.length + " bytes");
		check(fields.get(0).equals("Name"), "first field=" + fields.get(0));
		check(Collections.frequency(fields, "Tags") == 1000, "tag count=" + Collections.frequency(fields, "Tags"));

		System.out.println("OK consumedAtFirst=" + consumedAtFirst[0] + " total=" + data.length);
	}
}
`
	checkJava(t)

	buildAndRunJavaTest(t, "streamfeed", &javaSource{
		FileName: "IncrementalMain",
		Source:   incrementalHarness,
	})
}

// TestPullStream_Adapter exercises the virtual-thread pull adapter over the push
// streamer: full for-each iteration, the Stream() API, and early close.
func TestPullStream_Adapter(t *testing.T) {
	const pullHarness = `import com.opencost.streamfeed.Feed;
import com.opencost.streamfeed.FeedEncoder;
import com.opencost.streamfeed.FeedStreamer;
import com.bingen.BingenElement;
import com.bingen.BingenPullStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public final class PullMain {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) {
        byte[] data = FeedEncoder.toBytes(new Feed("news", List.of("a", "b", "c"), 7, Map.of("k", "v")));

        // Full pull iteration via for-each.
        List<String> fields = new ArrayList<>();
        try (BingenPullStream pull = BingenPullStream.of(data, FeedStreamer::stream)) {
            for (BingenElement e : pull) {
                fields.add(e.field().name());
            }
        }
        check(fields.equals(List.of("Name", "Tags", "Tags", "Tags", "Count", "Meta")), "fields=" + fields);

        // Stream() API, collecting the flattened Tags element values.
        List<Object> tags = new ArrayList<>();
        try (BingenPullStream pull = BingenPullStream.of(data, FeedStreamer::stream)) {
            pull.stream()
                .filter(e -> e.field().name().equals("Tags"))
                .forEach(e -> tags.add(e.value().value()));
        }
        check(tags.equals(List.of("a", "b", "c")), "tags=" + tags);

        // Early close via break must not hang (producer virtual thread interrupted).
        List<String> partial = new ArrayList<>();
        try (BingenPullStream pull = BingenPullStream.of(data, FeedStreamer::stream)) {
            for (BingenElement e : pull) {
                partial.add(e.field().name());
                if (partial.size() == 2) {
                    break;
                }
            }
        }
        check(partial.size() == 2, "partial=" + partial);

        System.out.println("OK");
    }
}
`
	const goPackage = "streamfeed"
	const harnessName = "PullMain"

	checkJava(t)

	buildAndRunJavaTest(t, goPackage, &javaSource{
		FileName: harnessName,
		Source:   pullHarness,
	})
}
