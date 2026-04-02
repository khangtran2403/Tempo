import React, { useCallback, useState } from 'react';
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
  Controls,
  Background,
  MiniMap,
  NodeTypes,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { WorkflowDefinition } from '../../types/workflow';
import TriggerNode from './nodes/TriggerNode';
import ActionNode from './nodes/ActionNode';
import NodeSidebar from './NodeSidebar';
import NodeConfigModal from './NodeConfigModal';


const nodeTypes: NodeTypes = {
  trigger: TriggerNode,
  action: ActionNode,
};

interface WorkflowBuilderProps {
  initialDefinition?: WorkflowDefinition;
  onSave: (definition: WorkflowDefinition) => void;
}

export default function WorkflowBuilder({ initialDefinition, onSave }: WorkflowBuilderProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [configModalOpen, setConfigModalOpen] = useState(false);

  
  const loadDefinition = useCallback((definition: WorkflowDefinition) => {
    const newNodes: Node[] = [];
    const newEdges: Edge[] = [];

   
    newNodes.push({
      id: 'trigger',
      type: 'trigger',
      position: { x: 250, y: 50 },
      data: {
        label: definition.trigger.type,
        config: definition.trigger.config,
      },
    });

    
    definition.actions.forEach((action, index) => {
      newNodes.push({
        id: action.id,
        type: 'action',
        position: { x: 250, y: 200 + index * 150 },
        data: {
          label: action.type,
          config: action.config,
        },
      });

      
      const sourceId = index === 0 ? 'trigger' : definition.actions[index - 1].id;
      newEdges.push({
        id: `${sourceId}-${action.id}`,
        source: sourceId,
        target: action.id,
        animated: true,
      });
    });

    setNodes(newNodes);
    setEdges(newEdges);
  }, [setNodes, setEdges]);

  React.useEffect(() => {
    if (initialDefinition) {
      loadDefinition(initialDefinition);
    }
  }, [initialDefinition, loadDefinition]);

  
  const exportDefinition = (): WorkflowDefinition => {
    const triggerNode = nodes.find(n => n.id === 'trigger');
    const actionNodes = nodes.filter(n => n.type === 'action');

    return {
      trigger: {
        id: 'trigger',
        type: triggerNode?.data.label || 'webhook',
        config: triggerNode?.data.config || {},
      },
      actions: actionNodes.map(node => ({
        id: node.id,
        type: node.data.label,
        config: node.data.config,
      })),
    };
  };

  
  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds) => addEdge({ ...connection, animated: true }, eds));
    },
    [setEdges]
  );

  
  const addNode = (type: string, nodeType: 'trigger' | 'action') => {
    const id = nodeType === 'trigger' ? 'trigger' : `action_${Date.now()}`;
    const newNode: Node = {
      id,
      type: nodeType,
      position: { x: 250, y: 200 + nodes.length * 150 },
      data: {
        label: type,
        config: {},
      },
    };

    setNodes((nds) => [...nds, newNode]);
  };

 
  const openNodeConfig = (node: Node) => {
    setSelectedNode(node);
    setConfigModalOpen(true);
  };

  
  const updateNodeConfig = (data: { id: string; config: Record<string, any> }) => {
    if (!selectedNode) return;

    const { id: newId, config } = data;
    const oldId = selectedNode.id;

    setNodes((nds) =>
      nds.map((node) => {
        if (node.id === oldId) {
          return { ...node, id: newId, data: { ...node.data, config } };
        }
        return node;
      })
    );

    if (oldId !== newId) {
      setEdges((eds) =>
        eds.map((edge) => {
          const newEdge = { ...edge };
          if (newEdge.source === oldId) {
            newEdge.source = newId;
          }
          if (newEdge.target === oldId) {
            newEdge.target = newId;
          }
          newEdge.id = `${newEdge.source}-${newEdge.target}`;
          return newEdge;
        })
      );
    }

    setConfigModalOpen(false);
    setSelectedNode(null);
  };

 
  const handleSave = () => {
    const definition = exportDefinition();
    onSave(definition);
  };

  return (
    <div className="h-full flex">
      <div className="flex-1 bg-gray-50">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeDoubleClick={(_, node) => openNodeConfig(node)}
          nodeTypes={nodeTypes}
          fitView
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </div>

      <NodeSidebar onAddNode={addNode} onSave={handleSave} />

      {selectedNode && (
        <NodeConfigModal
          open={configModalOpen}
          node={selectedNode}
          onClose={() => setConfigModalOpen(false)}
          onSave={updateNodeConfig}
        />
      )}
    </div>
  );
}