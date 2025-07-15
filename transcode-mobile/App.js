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
  const [inputURL, setInputURL] = useState(
    "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_1MB.mp4"
  );

  const [resolutions, setResolutions] = useState({
    "144p": true,
    "240p": false,
    "360p": true,
    "480p": false,
    "720p": true,
    "1080p": false,
  });

  const [codec, setCodec] = useState("h264");
  const [submitting, setSubmitting] = useState(false);
  const [jobs, setJobs] = useState([]);

  useEffect(() => {
    loadJobs();
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

    const payload = {
      input_url: inputURL,
      resolutions: selected,
      codec,
      stream_name: "big-bunny-1080p",
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
    await Clipboard.setStringAsync(url);
    if (Platform.OS === 'android') {
      ToastAndroid.show("📋 MPD URL copied", ToastAndroid.SHORT);
    } else {
      Alert.alert("Copied", "MPD URL copied to clipboard");
    }
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
      {Object.keys(resolutions).map((res) => (
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
        {["h264", "hevc", "vvc"].map((opt) => (
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

      <View style={styles.submitBtn}>
        <Button title={submitting ? "Submitting..." : "Submit"} onPress={handleSubmit} disabled={submitting} />
      </View>

      <Text style={styles.label}>Recent Jobs:</Text>
      {jobs.map((job) => (
        <View key={job.job_id} style={styles.jobCard}>
          <Text style={styles.jobText}>📦 {job.job_id}</Text>
          <Text>📺 {job.stream_name}</Text>
          <Text>📹 {job.codec.toUpperCase()} → {job.required_resolutions || "N/A"}</Text>
          <TouchableOpacity onPress={() => copyToClipboard(job.mpd_url)}>
            <Text style={styles.mpdUrl}>🔗 {job.mpd_url}</Text>
          </TouchableOpacity>
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
