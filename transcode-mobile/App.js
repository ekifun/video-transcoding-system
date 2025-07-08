import React, { useState } from 'react';
import { View, Text, TextInput, Button, StyleSheet, ScrollView, Alert } from 'react-native';
import Checkbox from 'expo-checkbox';

export default function App() {
  const [inputURL, setInputURL] = useState(
    "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_1MB.mp4"
  );
  const [resolutions, setResolutions] = useState({
    "144p": true,
    "360p": true,
    "720p": true,
  });
  const [submitting, setSubmitting] = useState(false);

  const handleCheckboxChange = (key) => {
    setResolutions((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    const selected = Object.keys(resolutions).filter((res) => resolutions[res]);

    const payload = {
      input_url: inputURL,
      resolutions: selected,
      codec: "h264",
    };

    try {
      const res = await fetch("http://13.57.143.121:8080/transcode", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (res.ok) {
        Alert.alert("✅ Job Submitted", `Job ID: ${data.job_id}`);
      } else {
        Alert.alert("❌ Submission Failed", JSON.stringify(data));
      }
    } catch (err) {
      Alert.alert("❌ Error", err.message);
    } finally {
      setSubmitting(false);
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

      <View style={styles.submitBtn}>
        <Button title={submitting ? "Submitting..." : "Submit"} onPress={handleSubmit} disabled={submitting} />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    padding: 20,
    paddingTop: 50,
  },
  title: {
    fontSize: 22,
    fontWeight: "bold",
    marginBottom: 20,
    textAlign: "center",
  },
  label: {
    fontWeight: "bold",
    marginTop: 20,
  },
  input: {
    borderColor: "#999",
    borderWidth: 1,
    padding: 10,
    borderRadius: 5,
    marginTop: 5,
    backgroundColor: "#fff",
  },
  checkboxRow: {
    flexDirection: "row",
    alignItems: "center",
    marginVertical: 5,
  },
  checkboxLabel: {
    marginLeft: 10,
  },
  submitBtn: {
    marginTop: 30,
  },
});
