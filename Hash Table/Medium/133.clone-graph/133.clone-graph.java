/*
// Definition for a Node.
class Node {
    public int val;
    public List<Node> neighbors;
    public Node() {
        val = 0;
        neighbors = new ArrayList<Node>();
    }
    public Node(int _val) {
        val = _val;
        neighbors = new ArrayList<Node>();
    }
    public Node(int _val, ArrayList<Node> _neighbors) {
        val = _val;
        neighbors = _neighbors;
    }
}
*/

class Solution {
    Map<Node, Node>cache;
    public Node cloneGraph(Node node) {
        cache=new HashMap<>();
        return dfs(node);
    }

    Node dfs(Node node){
        if(node==null) return node;
        if(cache.containsKey(node)) return cache.get(node);
        else{
            Node copy=new Node(node.val);
            cache.put(node, copy);
            for(Node nei:node.neighbors){
                copy.neighbors.add(dfs(nei));
            }
            return copy;
        }
    }
}