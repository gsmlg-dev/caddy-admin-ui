"use client";

import { useState } from "react";

export const Adapt = (props) => {
  const [json, setJson] = useState(null);

  return (
    <div className="card shadow-lg shadow-primary m-4">
      <div className="card-body">
        <h3 className="card-title">
          Convert <code>Caddyfile</code> to <code>JSON</code>
        </h3>
        <div className="w-full flex flex-row gap-4">
          <form
            className="flex flex-col gap-4 w-full"
            onSubmit={async (evt) => {
              evt.preventDefault();
              const data = new FormData(evt.target);
              const r = new Request("/adapt", {
                method: "POST",
                headers: {
                  "Content-Type": "text/caddyfile",
                },
                body: data.get("config"),
              });
              const resp = await fetch(r);
              const jsonData = await resp.json();
              setJson(jsonData);
            }}
          >
            <fieldset>
              <legend>Put your Caddyfile here</legend>
              <textarea
                name="config"
                className="textarea textarea-primary w-full min-h-64"
              ></textarea>
            </fieldset>
            <div className="flex gap-4">
              <button className="btn btn-primary btn-sm rounded-xl">
                Convert
              </button>
            </div>
          </form>
          <div className="flex flex-col w-full">
            <h3>JSON config</h3>
            <textarea
              className="textarea textarea-primary w-full min-h-64"
              value={
                json?.result != null ? JSON.stringify(json.result, null, 2) : ""
              }
              readOnly
            ></textarea>
          </div>
        </div>
      </div>
    </div>
  );
};
