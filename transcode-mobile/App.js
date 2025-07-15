<Text style={styles.label}>Codec:</Text>
<View style={styles.codecOptions}>
  {["h264", "hevc", "vvc", "vp9"].map((opt) => (
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
