# Update Protocol - Project-Memory

> **📝 MANDATORY PROCEDURES**: All Project-Memory updates must follow these protocols to maintain accuracy and consistency.

---

## Overview

Project-Memory documents are living records of the project's current state. They must be updated systematically to remain valuable. This document defines when, how, and who updates Project-Memory.

---

## Update Triggers

### Mandatory Update Triggers

**After Milestones**:
- Feature completion
- Version releases
- Major refactoring
- Architecture changes

**After Architectural Decisions**:
- New ADR created
- Design patterns adopted
- Technology choices made
- Infrastructure changes

**Monthly Reviews**:
- Review all Project-Memory documents
- Update metrics and status
- Archive obsolete information
- Identify documentation gaps

### Optional Update Triggers

**After Bug Fixes** (if significant):
- Update known issues
- Document lessons learned
- Update troubleshooting guides

**After Configuration Changes**:
- Update technical architecture
- Document configuration patterns

---

## Update Responsibilities

### Primary Responsibility: Memory Specialist

The Memory Specialist coordinates all Project-Memory updates:
- Identifies when updates needed
- Assigns content updates to specialists
- Reviews updates for accuracy
- Ensures consistency across documents
- Maintains document versions

### Contributing Specialists

**All Specialists** can contribute content:
- Go Specialist → Technical architecture (Go-specific)
- Test Specialist → Testing strategy, quality metrics
- Architecture Specialist → Architecture decisions, patterns
- Data Specialist → Data architecture, schemas
- Domain Specialists → Domain-specific content

### Review Process

1. **Content Provider**: Specialist drafts update
2. **Coordination**: Memory Specialist reviews
3. **Validation**: Orchestrator checks consistency
4. **Commit**: Memory Specialist applies update
5. **Notification**: Team notified of significant updates

---

## Document-Specific Update Guidelines

### project-overview.md

**Update Frequency**: After milestones, monthly minimum

**Sections to Update**:
- **Current Status**: Always current
- **Recent Milestones**: Add completed milestones
- **Active Work Streams**: Current priorities
- **Recent Decisions**: New ADRs
- **Key Metrics**: Monthly review
- **Known Issues**: As discovered/resolved

**Who Updates**:
- Memory Specialist (coordination)
- Orchestrator (status review)
- All specialists (their domains)

---

### technical-architecture.md

**Update Frequency**: After architectural changes

**Sections to Update**:
- **Architecture Diagrams**: After structural changes
- **Design Patterns**: New patterns added
- **Data Flow**: After routing changes
- **Technology Decisions**: New choices documented
- **Performance Considerations**: After optimizations

**Who Updates**:
- Architecture Specialist (lead)
- Go Specialist (implementation details)
- Memory Specialist (coordination)

---

### development-history.md

**Update Frequency**: Monthly, after milestones

**Sections to Update**:
- **Timeline**: Append new milestones
- **Version History**: Document releases
- **Major Decisions**: Link to ADRs
- **Team Evolution**: Role changes

**Who Updates**:
- Memory Specialist (primary author)
- All specialists (contributions)

---

### memory-management.md

**Update Frequency**: After memory system changes

**Sections to Update**:
- **Memory Operations**: New operations
- **Schemas**: Schema changes
- **Performance**: Memory metrics
- **Troubleshooting**: New issues/solutions

**Who Updates**:
- Memory Specialist (primary)
- Data Specialist (schemas)

---

## Update Process

### Standard Update Workflow

1. **Identify Need**
   - Trigger occurs (milestone, decision, etc.)
   - Memory Specialist creates update task
   - Task added to `Memory-System/short-term/active-tasks.yaml`

2. **Gather Content**
   - Relevant specialists provide updates
   - Context loaded from memory system
   - Previous versions reviewed

3. **Draft Update**
   - Specialist(s) draft changes
   - Follow document structure
   - Maintain consistent voice
   - Include examples where helpful

4. **Review**
   - Memory Specialist reviews for:
     - Accuracy
     - Completeness
     - Consistency with other docs
     - Proper formatting
   - Orchestrator validates routing and roles

5. **Apply Update**
   - Memory Specialist updates document(s)
   - Updates "Last Updated" metadata
   - Commits to version control
   - Updates changelog

6. **Notify**
   - Significant updates communicated to team
   - Related memory system files updated
   - Cross-references validated

---

## Quality Standards

### Content Quality

**All Updates Must**:
- Be accurate and current
- Use clear, concise language
- Follow document structure
- Include examples where helpful
- Maintain consistent terminology
- Link to related documents

**Avoid**:
- Outdated information
- Speculation or future plans without marking as such
- Inconsistencies with other documents
- Jargon without explanation
- Broken links

---

### Format Standards

**Markdown**:
- Use consistent heading levels
- Use tables for structured data
- Use code blocks for code/commands
- Use blockquotes for important notes
- Use lists for sequential/hierarchical info

**Metadata**:
- Include "Last Updated" date
- Include "Updated By" (specialist role)
- Include "Version" if applicable

**Links**:
- Link to related documents
- Link to code references
- Link to external resources
- Keep links current

---

## Validation Checklist

### Before Committing Updates

- [ ] Content is accurate and current
- [ ] Follows document structure
- [ ] Formatting consistent
- [ ] Links all work
- [ ] Cross-references validated
- [ ] Metadata updated
- [ ] No sensitive information exposed
- [ ] Spell-checked
- [ ] Grammar-checked
- [ ] Reviewed by Memory Specialist

---

## Emergency Updates

### Urgent Updates (Security, Critical Bugs)

**Process**:
1. Identify urgency
2. Fast-track review
3. Update immediately
4. Notify team
5. Full review post-update

**Examples**:
- Security vulnerabilities discovered
- Critical bug patterns identified
- Breaking changes deployed

---

## Version Control

### Git Practices

**Commit Messages**:
```
[Project-Memory] Update project-overview.md

- Added milestone: Enhanced Agent System
- Updated current status
- Added recent decision ADR-001
- Updated metrics

Updated by: Memory Specialist
```

**Branch Strategy**:
- Direct commits to main for routine updates
- Feature branches for major restructuring
- Pull requests for significant changes

---

## Archival

### When to Archive

**Move to Archive When**:
- Information obsolete
- Historic value only
- Replaced by newer content

**Archive Location**:
- `Project-Memory/archive/YYYY-MM/`
- Maintain structure
- Compress if large

**Archive Metadata**:
- Reason for archival
- Date archived
- Replacement document (if any)

---

## Metrics

### Update Metrics to Track

**Currency**:
- Days since last update per document
- Percentage of documents up-to-date

**Activity**:
- Updates per month
- Contributors per document
- Review cycles per update

**Quality**:
- Broken links count
- Review feedback
- Accuracy issues reported

---

## Common Pitfalls

### Pitfall: Stale Information

**Problem**: Documents not updated regularly  
**Solution**: 
- Mandatory monthly review
- Automated reminders
- Update triggers enforced

### Pitfall: Inconsistent Information

**Problem**: Contradictions between documents  
**Solution**:
- Cross-reference validation
- Memory Specialist review
- Single source of truth per topic

### Pitfall: Too Much Detail

**Problem**: Documents become unwieldy  
**Solution**:
- High-level in Project-Memory
- Details in code/long-term memory
- Link to detail sources

### Pitfall: Update Friction

**Problem**: Updates feel like burden  
**Solution**:
- Make updates part of workflow
- Templates for common updates
- Celebrate good documentation

---

## Templates

### Milestone Update Template

```markdown
## Milestone: [Name]

**Date**: YYYY-MM-DD  
**Status**: Completed / In Progress  
**Type**: Feature / Infrastructure / Refactoring

**Description**: [Brief description of milestone]

**Deliverables**:
- [ ] Deliverable 1
- [ ] Deliverable 2

**Impact**: [Impact on project]

**Related Decisions**: [Link to ADRs]
```

### Decision Update Template

```markdown
## ADR-XXX: [Decision Title]

**Date**: YYYY-MM-DD  
**Status**: Accepted / Rejected / Deprecated  
**Context**: [Why decision needed]  
**Decision**: [What was decided]  
**Consequences**: [Impact of decision]  
**Alternatives**: [What else was considered]
```

---

## Summary

Effective Project-Memory updates require:
- **Discipline**: Follow update triggers
- **Collaboration**: All specialists contribute
- **Quality**: High standards maintained
- **Consistency**: Unified voice and structure
- **Currency**: Regular reviews and updates

By maintaining current, accurate Project-Memory, the team has a reliable single source of truth about the project's state.

---

**Last Updated**: 2026-02-25  
**Updated By**: Memory Specialist  
**Version**: 1.0.0
