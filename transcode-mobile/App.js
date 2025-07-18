import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  Button,
  StyleSheet,
  ScrollView,
  Alert,
  TouchableOpacity,
  Platform,
  ToastAndroid,
} from 'react-native';
import Checkbox from 'expo-checkbox';
import * as Clipboard from 'expo-clipboard';

export default function App() {
  const RESOLUTION_OPTIONS = {
    "144p": false,
    "240p": false,
    "360p": false,
    "480p": false,
    "720p": false,
    "1080p": false,
  };

  const [inputURL, setInputURL] = useState(
    "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_1MB.mp4"
  );

  const [resolutions, setResolutions] = useState(RESOLUTION_OPTIONS);
  const [codec, setCodec] = useState("h264");
  const [gopSize, setGopSize] = useState("48");
  const [keyintMin, setKeyintMin] = useState("48");
  const [submitting, setSubmitting] = useState(false);
  const [jobs, setJobs] = useState([]);

  useEffect(() => {
    loadJobs();
    const interval = setInterval(() => {
      loadJobs();
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const loadJobs = async () => {
    try {
      const res = await fetch("http://13.57.143.121:8080/jobs");
      const data = await res.json();
      if (Array.isArray(data)) {
        setJobs(data);
      }
    } catch (err) {
      console.error("❌ Failed to load jobs:", err);
    }
  };

  const handleCheckboxChange = (key) => {
    setResolutions((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    const selected = Object.keys(resolutions).filter((res) => resolutions[res]);
    if (selected.length === 0) {
      Alert.alert("⚠️ Please select at least one resolution.");
      setSubmitting(false);
      return;
    }

    const payload = {
      input_url: inputURL,
      resolutions: selected,
      codec,
      stream_name: "big-bunny-1080p",
      gop_size: parseInt(gopSize) || 48,
      keyint_min: parseInt(keyintMin) || 48,
    };

    try {
      const res = await fetch("http://13.57.143.121:8080/transcode", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (res.ok) {
        Alert.alert("✅ Job Submitted", `Job ID: ${data.job_id}`);
        loadJobs();
      } else {
        Alert.alert("❌ Submission Failed", JSON.stringify(data));
      }
    } catch (err) {
      Alert.alert("❌ Error", err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const copyToClipboard = async (url) => {
    if (!url) return;
    await Clipboard.setStringAsync(url);
    if (Platform.OS === 'android') {
      ToastAndroid.show("📋 MPD URL copied", ToastAndroid.SHORT);
    } else {
      Alert.alert("Copied", "MPD URL copied to clipboard");
    }
  };

  const renderStatus = (status) => {
    let color = "#555";
    if (status === "waiting") color = "#999";
    if (status === "processing") color = "orange";
    if (status === "ready_for_mpd") color = "blue";
    if (status === "done") color = "green";

    return (
      <Text style={{ color, fontWeight: "bold" }}>
        {status ? status.toUpperCase() : "UNKNOWN"}
      </Text>
    );
  };

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>Transcode Job Submission</Text>

      <Text style={styles.label}>Input URL:</Text>
      <TextInput
        style={styles.input}
        value={inputURL}
        onChangeText={setInputURL}
        multiline
      />

      <Text style={styles.label}>Resolutions:</Text>
      {Object.keys(RESOLUTION_OPTIONS).map((res) => (
        <View key={res} style={styles.checkboxRow}>
          <Checkbox
            value={resolutions[res]}
            onValueChange={() => handleCheckboxChange(res)}
          />
          <Text style={styles.checkboxLabel}>{res}</Text>
        </View>
      ))}

      <Text style={styles.label}>Codec:</Text>
      <View style={styles.codecOptions}>
        {["h264", "hevc", "vvc", "vp9", "av1"].map((opt) => (
          <TouchableOpacity
            key={opt}
            onPress={() => setCodec(opt)}
            style={styles.radioRow}
          >
            <View style={styles.radioCircle}>
              {codec === opt && <View style={styles.radioDot} />}
            </View>
            <Text style={styles.checkboxLabel}>{opt.toUpperCase()}</Text>
          </TouchableOpacity>
        ))}
      </View>

      <Text style={styles.label}>GOP Size (-g):</Text>
      <TextInput
        style={styles.input}
        value={gopSize}
        onChangeText={setGopSize}
        keyboardType="numeric"
      />

      <Text style={styles.label}>Key Frame Interval (keyint_min):</Text>
      <TextInput
        style={styles.input}
        value={keyintMin}
        onChangeText={setKeyintMin}
        keyboardType="numeric"
      />

      <View style={styles.submitBtn}>
        <Button title={submitting ? "Submitting..." : "Submit"} onPress={handleSubmit} disabled={submitting} />
      </View>

      <Text style={styles.label}>Recent Jobs (Auto-updating):</Text>
      {jobs.map((job) => (
        <View key={job.job_id} style={styles.jobCard}>
          <Text style={styles.jobText}>📦 {job.job_id}</Text>
          <Text>📺 {job.stream_name}</Text>
          <Text>📹 {job.codec.toUpperCase()} → {job.representations || "N/A"}</Text>
          <Text>Status: {renderStatus(job.status)}</Text>
          {job.mpd_url ? (
            <TouchableOpacity onPress={() => copyToClipboard(job.mpd_url)}>
              <Text style={styles.mpdUrl}>🔗 {job.mpd_url}</Text>
            </TouchableOpacity>
          ) : (
            <Text style={{ color: "#aaa" }}>⏳ MPD Not Available</Text>
          )}
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20, paddingTop: 50 },
  title: { fontSize: 22, fontWeight: "bold", marginBottom: 20, textAlign: "center" },
  label: { fontWeight: "bold", marginTop: 20 },
  input: { borderColor: "#999", borderWidth: 1, padding: 10, borderRadius: 5, marginTop: 5, backgroundColor: "#fff" },
  checkboxRow: { flexDirection: "row", alignItems: "center", marginVertical: 5 },
  checkboxLabel: { marginLeft: 10 },
  submitBtn: { marginTop: 30 },
  codecOptions: { marginTop: 10 },
  radioRow: { flexDirection: "row", alignItems: "center", marginVertical: 5 },
  radioCircle: {
    height: 20, width: 20, borderRadius: 10, borderWidth: 2,
    borderColor: "#555", alignItems: "center", justifyContent: "center", marginRight: 10
  },
  radioDot: { height: 10, width: 10, borderRadius: 5, backgroundColor: "#555" },
  jobCard: {
    marginTop: 15,
    padding: 10,
    backgroundColor: "#eef",
    borderRadius: 5,
    borderColor: "#ccd",
    borderWidth: 1,
  },
  jobText: {
    fontWeight: "bold",
    marginBottom: 4,
  },
  mpdUrl: {
    color: 'blue',
    marginTop: 5,
    textDecorationLine: 'underline',
    flexShrink: 1,
    flexWrap: 'wrap',
  },
});
