-- Instances no longer get an automatic lease check. Liveness comes from checks an agent defines.
DELETE FROM health_checks WHERE source = 'lease';
