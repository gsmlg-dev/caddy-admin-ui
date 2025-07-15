"use client";

import { useState, useEffect } from "react";

export const Metrics = ({}) => {
  const [data, setData] = useState([]);
  useEffect(() => {
    fetch("/metrics")
      .then((resp) => resp.text())
      .then((data) => setData(data));
  }, []);

  return (
    <div className="flex gap-4">
      <pre>{data}</pre>
    </div>
  );
};
