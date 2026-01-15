package main

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReadEnv(t *testing.T) {
	// Test case 1: Valid positive integer
	if err := os.Setenv("TEST_COUNT", "10"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("TEST_COUNT") }()
	result := readEnv("TEST_COUNT", 5)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}

	// Test case 2: Invalid string (should use default)
	if err := os.Setenv("TEST_COUNT", "invalid"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("TEST_COUNT") }()
	result = readEnv("TEST_COUNT", 5)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}

	// Test case 3: Zero value (should use default)
	if err := os.Setenv("TEST_COUNT", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("TEST_COUNT") }()
	result = readEnv("TEST_COUNT", 5)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}

	// Test case 4: Negative value (should use default)
	if err := os.Setenv("TEST_COUNT", "-1"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("TEST_COUNT") }()
	result = readEnv("TEST_COUNT", 5)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}

	// Test case 5: Empty string (should use default)
	if err := os.Setenv("TEST_COUNT", ""); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("TEST_COUNT") }()
	result = readEnv("TEST_COUNT", 5)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}

	// Test case 6: Environment variable not set (should use default)
	result = readEnv("NON_EXISTENT", 7)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}
}

func TestNamespace(t *testing.T) {
	// Test with env set
	if err := os.Setenv("NAMESPACE", "test-ns"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("NAMESPACE") }()
	actual := func(env string, defaultValue string) string {
		if value := os.Getenv(env); value != "" {
			return value
		}
		return defaultValue
	}("NAMESPACE", "logging")
	if actual != "test-ns" {
		t.Errorf("Expected test-ns, got %s", actual)
	}

	// Test with env not set
	_ = os.Unsetenv("NAMESPACE")
	actual = func(env string, defaultValue string) string {
		if value := os.Getenv(env); value != "" {
			return value
		}
		return defaultValue
	}("NAMESPACE", "logging")
	if actual != "logging" {
		t.Errorf("Expected logging, got %s", actual)
	}
}

func TestApiVKindName(t *testing.T) {
	defaultValue := []string{
		"integreatly.org/v1alpha1",
		"GrafanaDashboard",
		"graylog-grafana-dashboard-vm",
		"04e98ff7-7471-451f-a9cf-4bcad4a1bd41",
	}

	// Test with invalid length (4 instead of 5)
	if err := os.Setenv("INVOLVEDOBJECT", "a,b,c,d"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("INVOLVEDOBJECT") }()
	actual := func(env string, defaultValue []string) []string {
		result := strings.Split(os.Getenv(env), ",")
		if len(result) != 5 {
			return defaultValue
		}
		for _, res := range result {
			if res == "" {
				return defaultValue
			}
		}
		return result
	}("INVOLVEDOBJECT", defaultValue)
	if actual[0] != "integreatly.org/v1alpha1" {
		t.Errorf("Expected default, got %v", actual)
	}

	// Test with valid input
	if err := os.Setenv("INVOLVEDOBJECT", "v1,Kind,Name,UID,Version"); err != nil {
		t.Fatal(err)
	}
	actual = func(env string, defaultValue []string) []string {
		result := strings.Split(os.Getenv(env), ",")
		if len(result) != 5 {
			return defaultValue
		}
		for _, res := range result {
			if res == "" {
				return defaultValue
			}
		}
		return result
	}("INVOLVEDOBJECT", defaultValue)
	expected := []string{"v1", "Kind", "Name", "UID", "Version"}
	for i, v := range expected {
		if actual[i] != v {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	}

	// Test with empty value in split
	if err := os.Setenv("INVOLVEDOBJECT", "v1,,Name,UID,Version"); err != nil {
		t.Fatal(err)
	}
	actual = func(env string, defaultValue []string) []string {
		result := strings.Split(os.Getenv(env), ",")
		if len(result) != 5 {
			return defaultValue
		}
		for _, res := range result {
			if res == "" {
				return defaultValue
			}
		}
		return result
	}("INVOLVEDOBJECT", defaultValue)
	if actual[0] != "integreatly.org/v1alpha1" {
		t.Errorf("Expected default due to empty, got %v", actual)
	}
}

func TestCreateEvent(t *testing.T) {
	randomizer := rand.New(rand.NewSource(42)) // Fixed seed for deterministic test
	namespace := "test-namespace"
	apiVKindName := []string{"v1", "Pod", "test-pod", "uid-123", "version-1"}

	event := createEvent(1, randomizer, namespace, apiVKindName)

	if event.Namespace != namespace {
		t.Errorf("Expected namespace %s, got %s", namespace, event.Namespace)
	}

	if event.InvolvedObject.Kind != "Pod" {
		t.Errorf("Expected Kind Pod, got %s", event.InvolvedObject.Kind)
	}

	if event.InvolvedObject.Name != "test-pod" {
		t.Errorf("Expected Name test-pod, got %s", event.InvolvedObject.Name)
	}

	if event.Type != "Normal" {
		t.Errorf("Expected Type Normal, got %s", event.Type)
	}

	if event.Reason != "Completed" {
		t.Errorf("Expected Reason Completed, got %s", event.Reason)
	}

	if event.Message != "This is test message of Event to load cloud-events-reader. Do not worry" {
		t.Errorf("Unexpected message")
	}

	if event.Count != 1 {
		t.Errorf("Expected Count 1, got %d", event.Count)
	}

	if event.Source.Component != "k8s-event-generator" {
		t.Errorf("Expected Component k8s-event-generator, got %s", event.Source.Component)
	}

	// Check that timestamps are set (approximately now)
	now := time.Now()
	if now.Sub(event.FirstTimestamp.Time) > time.Minute {
		t.Errorf("FirstTimestamp not recent")
	}
	if now.Sub(event.LastTimestamp.Time) > time.Minute {
		t.Errorf("LastTimestamp not recent")
	}
}

func TestRunGenerator(t *testing.T) {
	// Create a fake Kubernetes client
	fakeClient := fake.NewClientset()

	namespace := "test-namespace"
	apiVKindName := []string{"v1", "Pod", "test-pod", "uid-123", "version-1"}
	count := 3
	sleep := 1
	maxLoops := 1 // Run only one loop for testing

	// Run the generator
	runGenerator(fakeClient, count, sleep, maxLoops, namespace, apiVKindName)

	// Verify that events were created
	events, err := fakeClient.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}

	if len(events.Items) != count {
		t.Errorf("Expected %d events, got %d", count, len(events.Items))
	}

	for i, event := range events.Items {
		if event.Namespace != namespace {
			t.Errorf("Event %d: expected namespace %s, got %s", i, namespace, event.Namespace)
		}
		if event.InvolvedObject.Kind != apiVKindName[1] {
			t.Errorf("Event %d: expected kind %s, got %s", i, apiVKindName[1], event.InvolvedObject.Kind)
		}
		if event.Type != "Normal" {
			t.Errorf("Event %d: expected type Normal, got %s", i, event.Type)
		}
		if event.Reason != "Completed" {
			t.Errorf("Event %d: expected reason Completed, got %s", i, event.Reason)
		}
	}
}
