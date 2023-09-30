ALTER TABLE internal.users
    ADD CONSTRAINT fk_organization_id FOREIGN KEY (organization_id) REFERENCES internal.organizations (id)