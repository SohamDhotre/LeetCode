/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public ListNode sortList(ListNode head) {
        if(head==null || head.next==null) return head;
        ListNode mid=getMid(head); // get mid prev and dis-connect the link to 
        // have seperate left and right lists
        ListNode right=mid.next;
        mid.next=null;
        ListNode left=sortList(head);
        right=sortList(right);
        // seperate until they have only 1 node then merge both left and right
        return merge(left, right);
    }

    ListNode getMid(ListNode head){
        ListNode slow=head, fast=head, prev = null;;
        while(fast!=null && fast.next!=null){
            prev=slow;
            slow=slow.next;
            fast=fast.next.next;
        }
        // get mid-1 to remove the link (mid-1).next=null
        return prev;
    }

    ListNode merge(ListNode left, ListNode right){
        ListNode dummy=new ListNode(0);
        ListNode cur=dummy;
        while(left!=null && right!=null){
            if(left.val<=right.val){
                cur.next=left;
                left=left.next;
            } else {
                cur.next=right;
                right=right.next;
            }
            cur=cur.next;
        }
        // as list have the pointers to next node in the list, 
        // we can append directly the complete list
        cur.next=(left!=null)?left:right;
        return dummy.next;
    }
}