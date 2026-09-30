// INTENTIONALLY VULNERABLE. Do not deploy. See README.md in this folder.
//
// This is the kind of Express backend an AI assistant produces from a prompt like
// "build me a notes API with login and an AI summary endpoint". Every flaw below
// is one that vibescan is expected to catch.

const express = require("express");
const cors = require("cors");
const jwt = require("jsonwebtoken");
const { Pool } = require("pg");
const { exec } = require("child_process");

const app = express();
app.use(express.json());

// V06: wildcard CORS with credentials
app.use(cors({ origin: "*", credentials: true }));

// V06: hardcoded signing secret and database password
const JWT_SECRET = "supersecret123";
const db = new Pool({
  connectionString: "postgres://postgres:postgres123@db.example.com:5432/notes",
});

// V06: default admin credentials, V08: no rate limiting on login
app.post("/api/login", async (req, res) => {
  const { email, password } = req.body;
  if (email === "admin@example.com" && password === "admin") {
    return res.json({ token: jwt.sign({ email, role: "admin" }, JWT_SECRET) });
  }
  // V05: SQL injection through string concatenation
  const result = await db.query(
    "SELECT * FROM users WHERE email = '" + email + "' AND password = '" + password + "'"
  );
  if (result.rows.length === 0) return res.status(401).json({ error: "invalid" });
  res.json({ token: jwt.sign({ id: result.rows[0].id }, JWT_SECRET) });
});

// V02: no authentication, and any note can be read by ID (IDOR)
app.get("/api/notes/:id", async (req, res) => {
  const result = await db.query(`SELECT * FROM notes WHERE id = ${req.params.id}`);
  res.json(result.rows[0]);
});

// V05: command injection
app.post("/api/export", (req, res) => {
  exec("zip -r /tmp/export.zip " + req.body.folder, (err, stdout) => {
    res.json({ ok: !err, output: stdout });
  });
});

// V05: server-side request forgery
app.get("/api/preview", async (req, res) => {
  const response = await fetch(req.query.url);
  res.send(await response.text());
});

// V09: user input in the system prompt, model output passed to eval
app.post("/api/summarize", async (req, res) => {
  const response = await fetch("https://api.example-llm.com/v1/chat", {
    method: "POST",
    body: JSON.stringify({
      system: "You are a helpful assistant. The user's name is " + req.body.name,
      prompt: req.body.text,
    }),
  });
  const data = await response.json();
  res.json({ summary: eval(data.output) });
});

// V05: reflected XSS
app.get("/search", (req, res) => {
  res.send("<h1>Results for " + req.query.q + "</h1>");
});

app.listen(3000);
