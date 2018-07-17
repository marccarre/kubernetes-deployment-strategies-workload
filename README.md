# kubernetes-deployment-strategies-workload

Sample workload for the demo available at [github.com/marccarre/kubernetes-deployment-strategies](https://github.com/marccarre/kubernetes-deployment-strategies).

Features:

- It stores & reads users.
- Data is persisted in a PostgreSQL database.
- Database schema is managed via migrations (see `./pkg/db/migrations`).
- `v1.1.0` is backward compatible with `v1.0.0`.
