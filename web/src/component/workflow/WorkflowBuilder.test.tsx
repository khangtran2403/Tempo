import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ReactFlowProvider } from 'reactflow';
import WorkflowBuilder from './WorkflowBuilder';
import { WorkflowDefinition } from '../../types/workflow';

// Mocks for child components are tricky with React Flow, so we'll test them via integration.
// We can mock the onSave function to see what the output is.

const renderComponent = (props: { initialDefinition?: WorkflowDefinition; onSave: (def: WorkflowDefinition) => void; }) => {
  return render(
    <ReactFlowProvider>
      <WorkflowBuilder {...props} />
    </ReactFlowProvider>
  );
};

describe('WorkflowBuilder', () => {
  it('renders the builder and its sidebar', () => {
    const handleSave = jest.fn();
    renderComponent({ onSave: handleSave });

    // Check for React Flow canvas
    expect(screen.getByTestId('rf__controls')).toBeInTheDocument(); // Controls are a good indicator

    // Check for sidebar elements
    expect(screen.getByText('Triggers')).toBeInTheDocument();
    expect(screen.getByText('Actions')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /save workflow/i })).toBeInTheDocument();
  });

  it('allows adding a trigger and an action node', async () => {
    const handleSave = jest.fn();
    renderComponent({ onSave: handleSave });

    // Add a webhook trigger
    const webhookButton = screen.getByRole('button', { name: /webhook/i });
    fireEvent.click(webhookButton);

    // Add an HTTP action
    const httpButton = screen.getByRole('button', { name: /http request/i });
    fireEvent.click(httpButton);

    // Wait for nodes to appear
    await waitFor(() => {
      // The trigger node should be rendered
      expect(screen.getByText('webhook')).toBeInTheDocument();
      // The action node should be rendered
      expect(screen.getByText('http')).toBeInTheDocument();
    });
  });

  it('calls onSave with the correct definition when save button is clicked', async () => {
    const handleSave = jest.fn();
    renderComponent({ onSave: handleSave });

    // Add nodes
    fireEvent.click(screen.getByRole('button', { name: /webhook/i }));
    fireEvent.click(screen.getByRole('button', { name: /http request/i }));

    // Wait for nodes to be in the state before saving
    await waitFor(() => {
      expect(screen.getByText('http')).toBeInTheDocument();
    });

    // Click save
    fireEvent.click(screen.getByRole('button', { name: /save workflow/i }));

    // Check the data passed to onSave
    expect(handleSave).toHaveBeenCalledTimes(1);
    const savedDefinition = handleSave.mock.calls[0][0] as WorkflowDefinition;
    
    // Check trigger
    expect(savedDefinition.trigger.type).toBe('webhook');
    
    // Check action
    expect(savedDefinition.actions).toHaveLength(1);
    expect(savedDefinition.actions[0].type).toBe('http');
  });

  it('loads an initial definition correctly', async () => {
    const handleSave = jest.fn();
    const initialDefinition: WorkflowDefinition = {
      trigger: { id: 'trigger', type: 'cron', config: { expression: '* * * * *' } },
      actions: [
        { id: 'action_1', type: 'email', config: { to: 'test@test.com' } },
      ],
    };

    renderComponent({ initialDefinition, onSave: handleSave });

    // Check that the nodes from the initial definition are rendered
    await waitFor(() => {
      expect(screen.getByText('cron')).toBeInTheDocument();
      expect(screen.getByText('email')).toBeInTheDocument();
    });
  });
});
