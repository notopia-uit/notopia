# Document replica event delivery

Keep the document service at one replica until event delivery and deletion races are fixed and
verified with multi-replica integration tests. Redis synchronization alone does not solve this.
The chart uses a Recreate deployment to avoid overlapping old and new pods during normal upgrades.

Document pods share a Kafka consumer group, so an event reaches one group member, not every pod.
Each pod owns its own WebSocket connections and loaded documents. Role-change and member-removal
handlers only update or close connections on the pod that receives the event. A user connected to
another pod may retain stale access until their connection is reauthenticated or closed.

When a note is deleted, the document service deletes its associated document row and revisions.
The receiving pod guards its loaded document against further stores, closes its local connections,
and waits for local pending stores before deleting the row. Another pod may miss the event and
subsequently upsert its loaded document, recreating the deleted row. The guard is local memory;
the Hocuspocus Redis extension does not broadcast these application event handlers.

Any fix must deliver revocations and deletion to all relevant live pods and coordinate deletion
with stores across replicas. Test deletion during active edits and in-flight stores, role changes,
member removal, pod restarts, and message delivery failures before allowing scale-out.
