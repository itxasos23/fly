package main

import (
  "encoding/json"
  "log"
  maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func contains(list []float64, value_to_find float64) bool {
  for _, value := range list {
    if value == value_to_find {
      return true
    }
  }
  return false
}

func main() {
  n := maelstrom.NewNode()

  seen_values := make([]float64, 0)
  var neighbors []string

  n.Handle("broadcast", func(msg maelstrom.Message) error {
    var request_body map[string]any
    if err := json.Unmarshal(msg.Body, &request_body); err != nil {
      return err
    }

    response_body := make(map[string]any)
    response_body["type"] = "broadcast_ok"
    n.Reply(msg, response_body)

    value := request_body["message"].(float64)
    if contains(seen_values, value) {
      return nil
    }

    seen_values = append(seen_values, value)

    propagation_body := make(map[string]any)
    propagation_body["type"] = "broadcast"
    propagation_body["message"] = value

    for idx := 0; idx < len(neighbors); idx++ {
      neighbor := neighbors[idx]
      if neighbor == n.ID() {continue}
      n.Send(neighbor, propagation_body)
    }

    return nil 
  })

  n.Handle("broadcast_ok", func(msg maelstrom.Message) error {return nil})

  n.Handle("read", func(msg maelstrom.Message) error {

    response_body := make(map[string]any)
    response_body["type"] = "read_ok"
    response_body["messages"] = &seen_values

    return n.Reply(msg, response_body)

  })

  n.Handle("topology", func(msg maelstrom.Message) error {
    var request_body map[string]any
    if err := json.Unmarshal(msg.Body, &request_body); err != nil {
      return err
    }

    topology := request_body["topology"].(map[string]any)
    new_neighbors_raw := topology[n.ID()].([]interface {})

    new_neighbors := make([]string, len(new_neighbors_raw))
    for i, v := range new_neighbors_raw {
        new_neighbors[i] = v.(string)
    }

    neighbors = append(neighbors, new_neighbors...) 

    response_body := make(map[string]any)
    response_body["type"] = "topology_ok"
    return n.Reply(msg, response_body)
  })


  if err := n.Run(); err != nil {
    log.Fatal(err)
  }
}
